package exporters

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gophercloud/gophercloud/v2/openstack/reservation/v1/hosts"
	"github.com/gophercloud/gophercloud/v2/openstack/reservation/v1/leases"
	"github.com/prometheus/client_golang/prometheus"
)

var knownLeaseStatuses = map[string]int{
	"CREATING":    0,
	"PENDING":     1,
	"STARTING":    2,
	"ACTIVE":      3,
	"UPDATING":    4,
	"TERMINATING": 5,
	"TERMINATED":  6,
	"DELETING":    7,
	"ERROR":       8,
}

func mapLeaseStatus(current string) int {
	return mapStatus(knownLeaseStatuses, current)
}

var knownReservationStatuses = map[string]int{
	"pending": 0,
	"active":  1,
	"deleted": 2,
	"error":   3,
}

func mapReservationStatus(current string) int {
	return mapStatus(knownReservationStatuses, current)
}

var (
	blazarHostLabels        = []string{"id", "hostname", "availability_zone", "hypervisor_type"}
	blazarLeaseLabels       = []string{"id", "name", "project_id"}
	blazarReservationLabels = []string{"id", "lease_id", "resource_type"}
)

type BlazarExporter struct {
	BaseOpenStackExporter
}

var defaultBlazarMetrics = []Metric{
	{Name: "total_hosts", Fn: ListBlazarHosts},
	{Name: "host_reservable", Labels: blazarHostLabels, Fn: nil},
	{Name: "host_vcpus", Labels: blazarHostLabels, Fn: nil},
	{Name: "host_memory_mb", Labels: blazarHostLabels, Fn: nil},
	{Name: "host_local_gb", Labels: blazarHostLabels, Fn: nil},
	{Name: "host_property", Labels: []string{"id", "hostname", "property", "value"}, Fn: nil},
	{Name: "host_allocated", Labels: blazarHostLabels, Fn: nil},
	{Name: "host_allocations", Labels: blazarHostLabels, Fn: nil},
	{Name: "reservation_allocated_hosts", Labels: blazarReservationLabels, Fn: nil},
	{Name: "total_leases", Fn: ListLeases},
	{Name: "lease_status", Labels: []string{"id", "name", "project_id", "user_id", "status"}, Fn: nil},
	{Name: "lease_status_counter", Labels: []string{"status"}, Fn: nil},
	{Name: "lease_degraded", Labels: blazarLeaseLabels, Fn: nil},
	{Name: "lease_start_date", Labels: blazarLeaseLabels, Fn: nil},
	{Name: "lease_end_date", Labels: blazarLeaseLabels, Fn: nil},
	{Name: "lease_event_time", Labels: []string{"lease_id", "event_type", "status"}, Fn: nil},
	{Name: "reservation_status", Labels: []string{"id", "lease_id", "resource_type", "status"}, Fn: nil},
	{Name: "reservation_missing_resources", Labels: blazarReservationLabels, Fn: nil},
	{Name: "reservation_resources_changed", Labels: blazarReservationLabels, Fn: nil},
	{Name: "reservation_amount", Labels: blazarReservationLabels, Fn: nil},
	{Name: "reservation_vcpus", Labels: blazarReservationLabels, Fn: nil},
	{Name: "reservation_memory_mb", Labels: blazarReservationLabels, Fn: nil},
	{Name: "reservation_disk_gb", Labels: blazarReservationLabels, Fn: nil},
	{Name: "reservation_min_hosts", Labels: blazarReservationLabels, Fn: nil},
	{Name: "reservation_max_hosts", Labels: blazarReservationLabels, Fn: nil},
}

func NewBlazarExporter(config *ExporterConfig, logger *slog.Logger) (*BlazarExporter, error) {
	exporter := BlazarExporter{
		BaseOpenStackExporter{
			Name:           "blazar",
			ExporterConfig: *config,
			logger:         logger,
		},
	}

	for _, metric := range defaultBlazarMetrics {
		if exporter.isDeprecatedMetric(&metric) {
			continue
		}
		if !exporter.isSlowMetric(&metric) {
			exporter.AddMetric(metric.Name, metric.Fn, metric.Labels, metric.DeprecatedVersion, nil)
		}
	}

	return &exporter, nil
}

func ListBlazarHosts(ctx context.Context, exporter *BaseOpenStackExporter, ch chan<- prometheus.Metric) error {
	allPagesHosts, err := hosts.List(exporter.ClientV2, hosts.ListOpts{}).AllPages(ctx)
	if err != nil {
		return err
	}

	allHosts, err := hosts.ExtractHosts(allPagesHosts)
	if err != nil {
		return err
	}

	allPagesAllocations, err := hosts.ListAllocations(exporter.ClientV2, hosts.ListAllocationsOpts{}).AllPages(ctx)
	if err != nil {
		return err
	}

	allAllocations, err := hosts.ExtractAllocations(allPagesAllocations)
	if err != nil {
		return err
	}

	// Allocations only hold the reservations of leases that have not ended.
	now := time.Now()
	hostAllocations := make(map[string][]hosts.AllocationReservation, len(allAllocations))
	reservationHosts := make(map[hosts.AllocationReservation]int)
	for _, allocation := range allAllocations {
		hostAllocations[allocation.ResourceID] = allocation.Reservations
		for _, reservation := range allocation.Reservations {
			reservationHosts[reservation]++
		}
	}

	ch <- prometheus.MustNewConstMetric(exporter.Metrics["total_hosts"].Metric,
		prometheus.GaugeValue, float64(len(allHosts)))

	for _, host := range allHosts {
		labels := []string{host.ID, host.HypervisorHostname, host.AvailabilityZone, host.HypervisorType}

		var reservable = 0
		if host.Reservable {
			reservable = 1
		}

		ch <- prometheus.MustNewConstMetric(exporter.Metrics["host_reservable"].Metric,
			prometheus.GaugeValue, float64(reservable), labels...)
		ch <- prometheus.MustNewConstMetric(exporter.Metrics["host_vcpus"].Metric,
			prometheus.GaugeValue, float64(host.VCPUs), labels...)
		ch <- prometheus.MustNewConstMetric(exporter.Metrics["host_memory_mb"].Metric,
			prometheus.GaugeValue, float64(host.MemoryMB), labels...)
		ch <- prometheus.MustNewConstMetric(exporter.Metrics["host_local_gb"].Metric,
			prometheus.GaugeValue, float64(host.LocalGB), labels...)

		for property, value := range host.ExtraCapabilities {
			ch <- prometheus.MustNewConstMetric(exporter.Metrics["host_property"].Metric,
				prometheus.GaugeValue, 1, host.ID, host.HypervisorHostname, property, fmt.Sprintf("%v", value))
		}

		var allocated = 0
		for _, reservation := range hostAllocations[host.ID] {
			if !now.Before(reservation.StartDate) && now.Before(reservation.EndDate) {
				allocated = 1
				break
			}
		}

		ch <- prometheus.MustNewConstMetric(exporter.Metrics["host_allocated"].Metric,
			prometheus.GaugeValue, float64(allocated), labels...)
		ch <- prometheus.MustNewConstMetric(exporter.Metrics["host_allocations"].Metric,
			prometheus.GaugeValue, float64(len(hostAllocations[host.ID])), labels...)
	}

	for reservation, count := range reservationHosts {
		ch <- prometheus.MustNewConstMetric(exporter.Metrics["reservation_allocated_hosts"].Metric,
			prometheus.GaugeValue, float64(count), reservation.ID, reservation.LeaseID, leases.ResourceTypeHost)
	}

	return nil
}

func ListLeases(ctx context.Context, exporter *BaseOpenStackExporter, ch chan<- prometheus.Metric) error {
	allPagesLeases, err := leases.List(exporter.ClientV2, leases.ListOpts{}).AllPages(ctx)
	if err != nil {
		return err
	}

	allLeases, err := leases.ExtractLeases(allPagesLeases)
	if err != nil {
		return err
	}

	leaseStatusCounter := make(map[string]int, len(knownLeaseStatuses))
	for k := range knownLeaseStatuses {
		leaseStatusCounter[k] = 0
	}

	ch <- prometheus.MustNewConstMetric(exporter.Metrics["total_leases"].Metric,
		prometheus.GaugeValue, float64(len(allLeases)))

	for _, lease := range allLeases {
		leaseStatusCounter[lease.Status]++

		// Blazar never purges ended leases, so they are only counted to keep
		// the number of series bounded.
		if lease.Status == "TERMINATED" {
			continue
		}

		var degraded = 0
		if lease.Degraded {
			degraded = 1
		}

		ch <- prometheus.MustNewConstMetric(exporter.Metrics["lease_status"].Metric,
			prometheus.GaugeValue, float64(mapLeaseStatus(lease.Status)),
			lease.ID, lease.Name, lease.ProjectID, lease.UserID, lease.Status)
		ch <- prometheus.MustNewConstMetric(exporter.Metrics["lease_degraded"].Metric,
			prometheus.GaugeValue, float64(degraded), lease.ID, lease.Name, lease.ProjectID)
		ch <- prometheus.MustNewConstMetric(exporter.Metrics["lease_start_date"].Metric,
			prometheus.GaugeValue, float64(lease.StartDate.Unix()), lease.ID, lease.Name, lease.ProjectID)
		ch <- prometheus.MustNewConstMetric(exporter.Metrics["lease_end_date"].Metric,
			prometheus.GaugeValue, float64(lease.EndDate.Unix()), lease.ID, lease.Name, lease.ProjectID)

		for _, event := range lease.Events {
			ch <- prometheus.MustNewConstMetric(exporter.Metrics["lease_event_time"].Metric,
				prometheus.GaugeValue, float64(event.Time.Unix()), lease.ID, event.EventType, event.Status)
		}

		for _, reservation := range lease.Reservations {
			labels := []string{reservation.ID, reservation.LeaseID, reservation.ResourceType}

			var missingResources = 0
			if reservation.MissingResources {
				missingResources = 1
			}

			var resourcesChanged = 0
			if reservation.ResourcesChanged {
				resourcesChanged = 1
			}

			ch <- prometheus.MustNewConstMetric(exporter.Metrics["reservation_status"].Metric,
				prometheus.GaugeValue, float64(mapReservationStatus(reservation.Status)),
				reservation.ID, reservation.LeaseID, reservation.ResourceType, reservation.Status)
			ch <- prometheus.MustNewConstMetric(exporter.Metrics["reservation_missing_resources"].Metric,
				prometheus.GaugeValue, float64(missingResources), labels...)
			ch <- prometheus.MustNewConstMetric(exporter.Metrics["reservation_resources_changed"].Metric,
				prometheus.GaugeValue, float64(resourcesChanged), labels...)

			// Blazar only returns the fields of the reservation's resource type.
			if reservation.Amount != nil {
				ch <- prometheus.MustNewConstMetric(exporter.Metrics["reservation_amount"].Metric,
					prometheus.GaugeValue, float64(*reservation.Amount), labels...)
			}
			if reservation.VCPUs != nil {
				ch <- prometheus.MustNewConstMetric(exporter.Metrics["reservation_vcpus"].Metric,
					prometheus.GaugeValue, float64(*reservation.VCPUs), labels...)
			}
			if reservation.MemoryMB != nil {
				ch <- prometheus.MustNewConstMetric(exporter.Metrics["reservation_memory_mb"].Metric,
					prometheus.GaugeValue, float64(*reservation.MemoryMB), labels...)
			}
			if reservation.DiskGB != nil {
				ch <- prometheus.MustNewConstMetric(exporter.Metrics["reservation_disk_gb"].Metric,
					prometheus.GaugeValue, float64(*reservation.DiskGB), labels...)
			}
			if reservation.Min != nil {
				ch <- prometheus.MustNewConstMetric(exporter.Metrics["reservation_min_hosts"].Metric,
					prometheus.GaugeValue, float64(*reservation.Min), labels...)
			}
			if reservation.Max != nil {
				ch <- prometheus.MustNewConstMetric(exporter.Metrics["reservation_max_hosts"].Metric,
					prometheus.GaugeValue, float64(*reservation.Max), labels...)
			}
		}
	}

	for status, count := range leaseStatusCounter {
		ch <- prometheus.MustNewConstMetric(exporter.Metrics["lease_status_counter"].Metric,
			prometheus.GaugeValue, float64(count), status)
	}

	return nil
}
