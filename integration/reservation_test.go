package integration

import (
	"testing"

	"github.com/gophercloud/gophercloud/v2/openstack/reservation/v1/leases"
	"github.com/openstack-exporter/openstack-exporter/integration/clients"
	"github.com/openstack-exporter/openstack-exporter/integration/funcs"
)

func TestReservationIntegration(t *testing.T) {
	clients.RequireLong(t)

	reservationClient, err := clients.NewReservationV1Client()
	if err != nil {
		t.Fatalf("Failed to build reservation client: %v", err)
	}

	computeClient, err := clients.NewComputeV2Client()
	if err != nil {
		t.Fatalf("Failed to build compute client: %v", err)
	}

	host, cleanupHost := funcs.RequireFreepoolHost(t, reservationClient, computeClient)
	defer cleanupHost()

	lease, err := funcs.CreateHostLease(t, reservationClient)
	if err != nil {
		t.Fatalf("Could not create test lease: %v", err)
	}
	defer funcs.DeleteLease(t, reservationClient, lease)
	reservation := lease.Reservations[0]

	cleanup := startExporter(t, "reservation")
	defer cleanup()
	metrics := scrapeLoggedMetrics(t, "")

	t.Run("openstack_blazar_up_metric", func(t *testing.T) {
		metrics.requireUp(t, "openstack_blazar_up")
	})

	t.Run("blazar_hosts_counted", func(t *testing.T) {
		metrics.requireMinValue(t, "openstack_blazar_total_hosts", nil, 1)
	})

	t.Run("blazar_host_metrics_present", func(t *testing.T) {
		metrics.requireMinValue(t, "openstack_blazar_host_vcpus", labels{"id": host.ID}, 1)
		metrics.requireMinValue(t, "openstack_blazar_host_memory_mb", labels{"id": host.ID}, 1)
		metrics.requireLabels(t, "openstack_blazar_host_reservable", labels{"id": host.ID}, "hostname")
	})

	t.Run("blazar_host_allocated_by_lease", func(t *testing.T) {
		metrics.requireMinValue(t, "openstack_blazar_host_allocations", labels{"id": host.ID}, 1)
		metrics.requireMinValue(t, "openstack_blazar_reservation_allocated_hosts", labels{
			"id":       reservation.ID,
			"lease_id": lease.ID,
		}, 1)
	})

	t.Run("blazar_lease_metrics_present", func(t *testing.T) {
		metrics.requireMinValue(t, "openstack_blazar_total_leases", nil, 1)
		metrics.requireLabels(t, "openstack_blazar_lease_status", labels{"id": lease.ID}, "name", "project_id", "status")
		metrics.requireMetric(t, "openstack_blazar_lease_degraded", labels{"id": lease.ID})
		metrics.requireMinValue(t, "openstack_blazar_lease_end_date", labels{"id": lease.ID}, 1)
		metrics.requireMinValue(t, "openstack_blazar_lease_event_time", labels{
			"lease_id":   lease.ID,
			"event_type": "end_lease",
		}, 1)
	})

	t.Run("blazar_reservation_metrics_present", func(t *testing.T) {
		match := labels{"id": reservation.ID, "lease_id": lease.ID, "resource_type": leases.ResourceTypeHost}
		metrics.requireLabels(t, "openstack_blazar_reservation_status", match, "status")
		metrics.requireMinValue(t, "openstack_blazar_reservation_min_hosts", match, 1)
		metrics.requireMinValue(t, "openstack_blazar_reservation_max_hosts", match, 1)
		metrics.requireNoMetric(t, "openstack_blazar_reservation_amount", match)
	})
}
