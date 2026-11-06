package funcs

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/hypervisors"
	"github.com/gophercloud/gophercloud/v2/openstack/reservation/v1/hosts"
	"github.com/gophercloud/gophercloud/v2/openstack/reservation/v1/leases"
	"github.com/openstack-exporter/openstack-exporter/integration/tools"
)

// RequireFreepoolHost makes sure the Blazar freepool holds a host to reserve.
// If it is empty, the first hypervisor is enrolled with a "gpu" capability.
// It returns the host and a function that removes the host if it was enrolled.
func RequireFreepoolHost(t *testing.T, client *gophercloud.ServiceClient, computeClient *gophercloud.ServiceClient) (*hosts.Host, func()) {
	allPages, err := hosts.List(client, hosts.ListOpts{}).AllPages(context.TODO())
	if err != nil {
		t.Fatalf("Unable to list Blazar hosts: %v", err)
	}

	allHosts, err := hosts.ExtractHosts(allPages)
	if err != nil {
		t.Fatalf("Unable to extract Blazar hosts: %v", err)
	}

	if len(allHosts) > 0 {
		t.Logf("Reusing freepool host: %s", allHosts[0].HypervisorHostname)
		return &allHosts[0], func() {}
	}

	allHypervisorPages, err := hypervisors.List(computeClient, nil).AllPages(context.TODO())
	if err != nil {
		t.Fatalf("Unable to list hypervisors: %v", err)
	}

	allHypervisors, err := hypervisors.ExtractHypervisors(allHypervisorPages)
	if err != nil {
		t.Fatalf("Unable to extract hypervisors: %v", err)
	}

	if len(allHypervisors) == 0 {
		t.Skip("no hypervisor available to enrol into the freepool")
	}

	host, err := hosts.Create(context.TODO(), client, hosts.CreateOpts{
		Name:              allHypervisors[0].HypervisorHostname,
		ExtraCapabilities: map[string]any{"gpu": "a100"},
	}).Extract()
	if err != nil {
		t.Fatalf("Unable to enrol hypervisor %s into the freepool: %v", allHypervisors[0].HypervisorHostname, err)
	}
	t.Logf("Created host: %s", host.ID)

	return host, func() {
		if err := hosts.Delete(context.TODO(), client, host.ID).ExtractErr(); err != nil {
			t.Fatalf("Unable to delete host %s: %v", host.ID, err)
		}
		t.Logf("Deleted host: %s", host.ID)
	}
}

// CreateHostLease creates a lease reserving one host from now on for an hour
// and waits for it to become active.
// An error will be returned if the lease was unable to be created.
func CreateHostLease(t *testing.T, client *gophercloud.ServiceClient) (*leases.Lease, error) {
	name := tools.RandomString("ACPTTEST", 16)
	t.Logf("Attempting to create lease: %s", name)

	lease, err := leases.Create(context.TODO(), client, leases.CreateOpts{
		Name:    name,
		EndDate: time.Now().UTC().Add(time.Hour),
		Reservations: []leases.ReservationOptsBuilder{
			leases.HostReservationOpts{Min: 1, Max: 1},
		},
	}).Extract()
	if err != nil {
		return nil, err
	}

	if len(lease.Reservations) != 1 {
		return nil, fmt.Errorf("lease %s holds %d reservations, expected 1", lease.ID, len(lease.Reservations))
	}

	// Blazar refuses to delete a lease while it is starting.
	if err := waitForStatus(lease.ID, "ACTIVE", func(ctx context.Context) (string, error) {
		latest, err := leases.Get(ctx, client, lease.ID).Extract()
		if err != nil {
			return "", err
		}
		return latest.Status, nil
	}); err != nil {
		return nil, err
	}

	t.Logf("Created lease: %s", lease.ID)
	return leases.Get(context.TODO(), client, lease.ID).Extract()
}

// DeleteLease deletes a lease via its ID and waits for it to be gone.
// A fatal error will occur if the lease failed to be deleted. This works
// best when using it as a deferred function.
func DeleteLease(t *testing.T, client *gophercloud.ServiceClient, lease *leases.Lease) {
	if err := leases.Delete(context.TODO(), client, lease.ID).ExtractErr(); err != nil {
		t.Fatalf("Unable to delete lease %s: %v", lease.ID, err)
	}

	// The host stays reserved until the lease is gone.
	if err := waitForDeleted(func(ctx context.Context) error {
		_, err := leases.Get(ctx, client, lease.ID).Extract()
		return err
	}); err != nil {
		t.Fatalf("Lease %s was not deleted: %v", lease.ID, err)
	}
	t.Logf("Deleted lease: %s", lease.ID)
}
