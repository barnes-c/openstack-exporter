package exporters

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

type BlazarTestSuite struct {
	BaseOpenStackTestSuite
}

var blazarExpectedUp = `
# HELP openstack_blazar_host_allocated host_allocated
# TYPE openstack_blazar_host_allocated gauge
openstack_blazar_host_allocated{availability_zone="nova",hostname="compute-1.example.com",hypervisor_type="QEMU",id="18"} 1
openstack_blazar_host_allocated{availability_zone="nova",hostname="compute-2.example.com",hypervisor_type="QEMU",id="19"} 0
# HELP openstack_blazar_host_allocations host_allocations
# TYPE openstack_blazar_host_allocations gauge
openstack_blazar_host_allocations{availability_zone="nova",hostname="compute-1.example.com",hypervisor_type="QEMU",id="18"} 2
openstack_blazar_host_allocations{availability_zone="nova",hostname="compute-2.example.com",hypervisor_type="QEMU",id="19"} 0
# HELP openstack_blazar_host_local_gb host_local_gb
# TYPE openstack_blazar_host_local_gb gauge
openstack_blazar_host_local_gb{availability_zone="nova",hostname="compute-1.example.com",hypervisor_type="QEMU",id="18"} 793
openstack_blazar_host_local_gb{availability_zone="nova",hostname="compute-2.example.com",hypervisor_type="QEMU",id="19"} 793
# HELP openstack_blazar_host_memory_mb host_memory_mb
# TYPE openstack_blazar_host_memory_mb gauge
openstack_blazar_host_memory_mb{availability_zone="nova",hostname="compute-1.example.com",hypervisor_type="QEMU",id="18"} 128297
openstack_blazar_host_memory_mb{availability_zone="nova",hostname="compute-2.example.com",hypervisor_type="QEMU",id="19"} 128297
# HELP openstack_blazar_host_property host_property
# TYPE openstack_blazar_host_property gauge
openstack_blazar_host_property{hostname="compute-1.example.com",id="18",property="gpu",value="a100"} 1
openstack_blazar_host_property{hostname="compute-1.example.com",id="18",property="rack",value="b12"} 1
# HELP openstack_blazar_host_reservable host_reservable
# TYPE openstack_blazar_host_reservable gauge
openstack_blazar_host_reservable{availability_zone="nova",hostname="compute-1.example.com",hypervisor_type="QEMU",id="18"} 1
openstack_blazar_host_reservable{availability_zone="nova",hostname="compute-2.example.com",hypervisor_type="QEMU",id="19"} 0
# HELP openstack_blazar_host_vcpus host_vcpus
# TYPE openstack_blazar_host_vcpus gauge
openstack_blazar_host_vcpus{availability_zone="nova",hostname="compute-1.example.com",hypervisor_type="QEMU",id="18"} 20
openstack_blazar_host_vcpus{availability_zone="nova",hostname="compute-2.example.com",hypervisor_type="QEMU",id="19"} 20
# HELP openstack_blazar_lease_degraded lease_degraded
# TYPE openstack_blazar_lease_degraded gauge
openstack_blazar_lease_degraded{id="7a9b1c2d-3e4f-4a5b-8c6d-7e8f9a0b1c2d",name="instance_lease",project_id="a2b6c11209974d6c916a190697d95f5f"} 1
openstack_blazar_lease_degraded{id="98c3544d-0afe-4251-8556-700c847a127f",name="host_lease",project_id="a2b6c11209974d6c916a190697d95f5f"} 0
openstack_blazar_lease_degraded{id="d3b8e1f2-4a5c-4b6d-9e7f-8a9b0c1d2e3f",name="future_host_lease",project_id="a2b6c11209974d6c916a190697d95f5f"} 0
# HELP openstack_blazar_lease_end_date lease_end_date
# TYPE openstack_blazar_lease_end_date gauge
openstack_blazar_lease_end_date{id="7a9b1c2d-3e4f-4a5b-8c6d-7e8f9a0b1c2d",name="instance_lease",project_id="a2b6c11209974d6c916a190697d95f5f"} 4.0760064e+09
openstack_blazar_lease_end_date{id="98c3544d-0afe-4251-8556-700c847a127f",name="host_lease",project_id="a2b6c11209974d6c916a190697d95f5f"} 4.0709088e+09
openstack_blazar_lease_end_date{id="d3b8e1f2-4a5c-4b6d-9e7f-8a9b0c1d2e3f",name="future_host_lease",project_id="a2b6c11209974d6c916a190697d95f5f"} 4.0760064e+09
# HELP openstack_blazar_lease_event_time lease_event_time
# TYPE openstack_blazar_lease_event_time gauge
openstack_blazar_lease_event_time{event_type="before_end_lease",lease_id="98c3544d-0afe-4251-8556-700c847a127f",status="UNDONE"} 4.0708224e+09
openstack_blazar_lease_event_time{event_type="end_lease",lease_id="7a9b1c2d-3e4f-4a5b-8c6d-7e8f9a0b1c2d",status="UNDONE"} 4.0760064e+09
openstack_blazar_lease_event_time{event_type="end_lease",lease_id="98c3544d-0afe-4251-8556-700c847a127f",status="UNDONE"} 4.0709088e+09
openstack_blazar_lease_event_time{event_type="end_lease",lease_id="d3b8e1f2-4a5c-4b6d-9e7f-8a9b0c1d2e3f",status="UNDONE"} 4.0760064e+09
openstack_blazar_lease_event_time{event_type="start_lease",lease_id="7a9b1c2d-3e4f-4a5b-8c6d-7e8f9a0b1c2d",status="UNDONE"} 4.0735872e+09
openstack_blazar_lease_event_time{event_type="start_lease",lease_id="98c3544d-0afe-4251-8556-700c847a127f",status="DONE"} 1.79395566e+09
openstack_blazar_lease_event_time{event_type="start_lease",lease_id="d3b8e1f2-4a5c-4b6d-9e7f-8a9b0c1d2e3f",status="UNDONE"} 4.0735872e+09
# HELP openstack_blazar_lease_start_date lease_start_date
# TYPE openstack_blazar_lease_start_date gauge
openstack_blazar_lease_start_date{id="7a9b1c2d-3e4f-4a5b-8c6d-7e8f9a0b1c2d",name="instance_lease",project_id="a2b6c11209974d6c916a190697d95f5f"} 4.0735872e+09
openstack_blazar_lease_start_date{id="98c3544d-0afe-4251-8556-700c847a127f",name="host_lease",project_id="a2b6c11209974d6c916a190697d95f5f"} 1.79395566e+09
openstack_blazar_lease_start_date{id="d3b8e1f2-4a5c-4b6d-9e7f-8a9b0c1d2e3f",name="future_host_lease",project_id="a2b6c11209974d6c916a190697d95f5f"} 4.0735872e+09
# HELP openstack_blazar_lease_status lease_status
# TYPE openstack_blazar_lease_status gauge
openstack_blazar_lease_status{id="7a9b1c2d-3e4f-4a5b-8c6d-7e8f9a0b1c2d",name="instance_lease",project_id="a2b6c11209974d6c916a190697d95f5f",status="PENDING",user_id="0a75eb6437d147a3bb7edf7aa4a18b79"} 1
openstack_blazar_lease_status{id="98c3544d-0afe-4251-8556-700c847a127f",name="host_lease",project_id="a2b6c11209974d6c916a190697d95f5f",status="ACTIVE",user_id="0a75eb6437d147a3bb7edf7aa4a18b79"} 3
openstack_blazar_lease_status{id="d3b8e1f2-4a5c-4b6d-9e7f-8a9b0c1d2e3f",name="future_host_lease",project_id="a2b6c11209974d6c916a190697d95f5f",status="PENDING",user_id="0a75eb6437d147a3bb7edf7aa4a18b79"} 1
# HELP openstack_blazar_lease_status_counter lease_status_counter
# TYPE openstack_blazar_lease_status_counter gauge
openstack_blazar_lease_status_counter{status="ACTIVE"} 1
openstack_blazar_lease_status_counter{status="CREATING"} 0
openstack_blazar_lease_status_counter{status="DELETING"} 0
openstack_blazar_lease_status_counter{status="ERROR"} 0
openstack_blazar_lease_status_counter{status="PENDING"} 2
openstack_blazar_lease_status_counter{status="STARTING"} 0
openstack_blazar_lease_status_counter{status="TERMINATED"} 1
openstack_blazar_lease_status_counter{status="TERMINATING"} 0
openstack_blazar_lease_status_counter{status="UPDATING"} 0
# HELP openstack_blazar_reservation_allocated_hosts reservation_allocated_hosts
# TYPE openstack_blazar_reservation_allocated_hosts gauge
openstack_blazar_reservation_allocated_hosts{id="c04d56e0-6b31-40b2-a57e-7283958100bd",lease_id="98c3544d-0afe-4251-8556-700c847a127f",resource_type="physical:host"} 1
openstack_blazar_reservation_allocated_hosts{id="e2a7c9d1-5b3f-4e8a-9c6d-1f2a3b4c5d6e",lease_id="d3b8e1f2-4a5c-4b6d-9e7f-8a9b0c1d2e3f",resource_type="physical:host"} 1
# HELP openstack_blazar_reservation_amount reservation_amount
# TYPE openstack_blazar_reservation_amount gauge
openstack_blazar_reservation_amount{id="5e1f3c0a-8f2d-4b6e-9a7c-3d4e5f6a7b8c",lease_id="7a9b1c2d-3e4f-4a5b-8c6d-7e8f9a0b1c2d",resource_type="virtual:instance"} 2
# HELP openstack_blazar_reservation_disk_gb reservation_disk_gb
# TYPE openstack_blazar_reservation_disk_gb gauge
openstack_blazar_reservation_disk_gb{id="5e1f3c0a-8f2d-4b6e-9a7c-3d4e5f6a7b8c",lease_id="7a9b1c2d-3e4f-4a5b-8c6d-7e8f9a0b1c2d",resource_type="virtual:instance"} 40
# HELP openstack_blazar_reservation_max_hosts reservation_max_hosts
# TYPE openstack_blazar_reservation_max_hosts gauge
openstack_blazar_reservation_max_hosts{id="c04d56e0-6b31-40b2-a57e-7283958100bd",lease_id="98c3544d-0afe-4251-8556-700c847a127f",resource_type="physical:host"} 2
openstack_blazar_reservation_max_hosts{id="e2a7c9d1-5b3f-4e8a-9c6d-1f2a3b4c5d6e",lease_id="d3b8e1f2-4a5c-4b6d-9e7f-8a9b0c1d2e3f",resource_type="physical:host"} 2
# HELP openstack_blazar_reservation_memory_mb reservation_memory_mb
# TYPE openstack_blazar_reservation_memory_mb gauge
openstack_blazar_reservation_memory_mb{id="5e1f3c0a-8f2d-4b6e-9a7c-3d4e5f6a7b8c",lease_id="7a9b1c2d-3e4f-4a5b-8c6d-7e8f9a0b1c2d",resource_type="virtual:instance"} 8192
# HELP openstack_blazar_reservation_min_hosts reservation_min_hosts
# TYPE openstack_blazar_reservation_min_hosts gauge
openstack_blazar_reservation_min_hosts{id="c04d56e0-6b31-40b2-a57e-7283958100bd",lease_id="98c3544d-0afe-4251-8556-700c847a127f",resource_type="physical:host"} 1
openstack_blazar_reservation_min_hosts{id="e2a7c9d1-5b3f-4e8a-9c6d-1f2a3b4c5d6e",lease_id="d3b8e1f2-4a5c-4b6d-9e7f-8a9b0c1d2e3f",resource_type="physical:host"} 1
# HELP openstack_blazar_reservation_missing_resources reservation_missing_resources
# TYPE openstack_blazar_reservation_missing_resources gauge
openstack_blazar_reservation_missing_resources{id="5e1f3c0a-8f2d-4b6e-9a7c-3d4e5f6a7b8c",lease_id="7a9b1c2d-3e4f-4a5b-8c6d-7e8f9a0b1c2d",resource_type="virtual:instance"} 1
openstack_blazar_reservation_missing_resources{id="c04d56e0-6b31-40b2-a57e-7283958100bd",lease_id="98c3544d-0afe-4251-8556-700c847a127f",resource_type="physical:host"} 0
openstack_blazar_reservation_missing_resources{id="e2a7c9d1-5b3f-4e8a-9c6d-1f2a3b4c5d6e",lease_id="d3b8e1f2-4a5c-4b6d-9e7f-8a9b0c1d2e3f",resource_type="physical:host"} 0
# HELP openstack_blazar_reservation_resources_changed reservation_resources_changed
# TYPE openstack_blazar_reservation_resources_changed gauge
openstack_blazar_reservation_resources_changed{id="5e1f3c0a-8f2d-4b6e-9a7c-3d4e5f6a7b8c",lease_id="7a9b1c2d-3e4f-4a5b-8c6d-7e8f9a0b1c2d",resource_type="virtual:instance"} 0
openstack_blazar_reservation_resources_changed{id="c04d56e0-6b31-40b2-a57e-7283958100bd",lease_id="98c3544d-0afe-4251-8556-700c847a127f",resource_type="physical:host"} 1
openstack_blazar_reservation_resources_changed{id="e2a7c9d1-5b3f-4e8a-9c6d-1f2a3b4c5d6e",lease_id="d3b8e1f2-4a5c-4b6d-9e7f-8a9b0c1d2e3f",resource_type="physical:host"} 0
# HELP openstack_blazar_reservation_status reservation_status
# TYPE openstack_blazar_reservation_status gauge
openstack_blazar_reservation_status{id="5e1f3c0a-8f2d-4b6e-9a7c-3d4e5f6a7b8c",lease_id="7a9b1c2d-3e4f-4a5b-8c6d-7e8f9a0b1c2d",resource_type="virtual:instance",status="pending"} 0
openstack_blazar_reservation_status{id="c04d56e0-6b31-40b2-a57e-7283958100bd",lease_id="98c3544d-0afe-4251-8556-700c847a127f",resource_type="physical:host",status="active"} 1
openstack_blazar_reservation_status{id="e2a7c9d1-5b3f-4e8a-9c6d-1f2a3b4c5d6e",lease_id="d3b8e1f2-4a5c-4b6d-9e7f-8a9b0c1d2e3f",resource_type="physical:host",status="pending"} 0
# HELP openstack_blazar_reservation_vcpus reservation_vcpus
# TYPE openstack_blazar_reservation_vcpus gauge
openstack_blazar_reservation_vcpus{id="5e1f3c0a-8f2d-4b6e-9a7c-3d4e5f6a7b8c",lease_id="7a9b1c2d-3e4f-4a5b-8c6d-7e8f9a0b1c2d",resource_type="virtual:instance"} 4
# HELP openstack_blazar_total_hosts total_hosts
# TYPE openstack_blazar_total_hosts gauge
openstack_blazar_total_hosts 2
# HELP openstack_blazar_total_leases total_leases
# TYPE openstack_blazar_total_leases gauge
openstack_blazar_total_leases 4
# HELP openstack_blazar_up up
# TYPE openstack_blazar_up gauge
openstack_blazar_up 1
`

func (suite *BlazarTestSuite) TestBlazarExporter() {
	err := testutil.CollectAndCompare(*suite.Exporter, strings.NewReader(blazarExpectedUp))
	assert.NoError(suite.T(), err)
}
