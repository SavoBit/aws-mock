/*
Copyright 2018 The Avi Networks.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package ec2

import (
	"strings"
	"time"

	randomdata "github.com/Pallinder/go-randomdata"
	ec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"

	"errors"
	"net"

	mock "github.com/stretchr/testify/mock"

	"github.com/satori/go.uuid"
)

// EC2API is an autogenerated mock type for the EC2API type
type EC2API struct {
	mock.Mock
	vpcs                   map[string]*ec2.Vpc
	vpcassocaiatedsubnet   map[string][]*ec2.Subnet         // vpc name will be key
	networkinterfaces      map[string]*ec2.NetworkInterface // interfaceid will be the name
	assignedIpOnSubnet     map[string][]string              // key subnet id
	subnets                map[string]*ec2.Subnet           // key subnet id
	assignedMacAddress     []string                         //assigned mac address
	assignedelasticIps     map[string]string                // key is allocation id
	assignedsecurityGroups map[string]*ec2.SecurityGroup    // key is security group id
	createdEc2instances    []*ec2.Instance
	defaultInstance        *ec2.Instance
	defaultSecurityGroupID string
	recorder               *Recorder
}

var AVI_STANDARD_ELASTIC_ALLOCATION_DOMAIN string = "aws"

var _ ec2iface.EC2API = &EC2API{}

//supported filters
var supportedsubnetfilter []string = []string{"vpc-id", "availabilityZone", "cidrBlock"}

func New() *EC2API {
	// aws allocate default security group to every instances
	securityGroupId, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	securityGroupIdStr := "sg-" + securityGroupId.String()
	securityGroupName := "sg-default"
	defaultSecurityGroup := &ec2.SecurityGroup{
		GroupId:   &securityGroupIdStr,
		GroupName: &securityGroupName,
	}
	instanceName := "avi-123"
	// TODO: @sch00lb0y find all the needed details and populate below
	defaultInstances := []*ec2.Instance{
		&ec2.Instance{
			InstanceId: &instanceName,
			SecurityGroups: []*ec2.GroupIdentifier{
				&ec2.GroupIdentifier{
					GroupId:   &securityGroupIdStr,
					GroupName: &securityGroupName,
				},
			},
		},
	}
	defaultSecurityGroups := map[string]*ec2.SecurityGroup{}
	defaultSecurityGroups[securityGroupIdStr] = defaultSecurityGroup
	recorder := &Recorder{}
	recorder.init()
	return &EC2API{
		vpcs:                   make(map[string]*ec2.Vpc, 0),
		vpcassocaiatedsubnet:   make(map[string][]*ec2.Subnet, 0),
		networkinterfaces:      make(map[string]*ec2.NetworkInterface, 0),
		assignedIpOnSubnet:     make(map[string][]string, 0),
		subnets:                make(map[string]*ec2.Subnet, 0),
		assignedMacAddress:     make([]string, 0),
		assignedelasticIps:     make(map[string]string, 0),
		assignedsecurityGroups: defaultSecurityGroups,
		createdEc2instances:    defaultInstances,
		defaultInstance:        defaultInstances[0],
		defaultSecurityGroupID: securityGroupIdStr,
		recorder:               recorder,
	}
}

func (_m *EC2API) EXPECT() *Recorder {
	return _m.recorder
}

func (_m *EC2API) AppendInstance(instance *ec2.Instance) {
	instance.SecurityGroups = append(instance.SecurityGroups, _m.defaultInstance.SecurityGroups...)
	_m.createdEc2instances = append(_m.createdEc2instances, instance)
}

func (_m *EC2API) AppendVpcs(vpc *ec2.Vpc) {
	_m.vpcs[*vpc.VpcId] = vpc
}

func (_m *EC2API) GetDefaultSecurityGroupID() string {
	return _m.defaultSecurityGroupID
}

// DescribeVpcs provides a mock function with given fields: _a0
func (_m *EC2API) DescribeVpcs(req *ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
	output := &ec2.DescribeVpcsOutput{}
	if err := _m.recorder.CheckError("DescribeVpcs"); err != nil {

		return output, err
	}
	_m.recorder.Record("DescribeVpcs")
	returns, exist := _m.recorder.giveRecordedOutput("DescribeVpcs", req)
	if exist {

		assertedErr, _ := returns[1].(error)
		return returns[0].(*ec2.DescribeVpcsOutput), assertedErr
	}
	for _, reqvpcid := range req.VpcIds {
		for _, vpc := range _m.vpcs {
			if reqvpcid == vpc.VpcId {
				output.Vpcs = append(output.Vpcs, vpc)
			}
		}
	}
	return output, nil
}

// CreateSubnet provides a mock function with given fields: _a0
func (_m *EC2API) CreateSubnet(_a0 *ec2.CreateSubnetInput) (output *ec2.CreateSubnetOutput, err error) {
	output = &ec2.CreateSubnetOutput{}
	if err := _m.recorder.CheckError("CreateSubnet"); err != nil {
		return output, err
	}
	_m.recorder.Record("CreateSubnet")
	returns, exist := _m.recorder.giveRecordedOutput("CreateSubnet", _a0)
	if exist {

		assertedErr, _ := returns[1].(error)
		return returns[0].(*ec2.CreateSubnetOutput), assertedErr
	}
	_, _, err = net.ParseCIDR(*_a0.CidrBlock)
	if err != nil {
		return nil, err
	}
	_, ok := _m.vpcs[*_a0.VpcId]
	if !ok {
		return nil, errors.New("vpc not exist")
	}
	subnetId := "subnet-" + rangeIn(1000, 9999)
	ipv6block := []*ec2.SubnetIpv6CidrBlockAssociation{}
	if _a0.Ipv6CidrBlock != nil {
		ipv6associationid := "subnet-cidr-assoc-" + rangeIn(1000, 9999)
		defaultState := "ASSOCIATED"
		ipv6block = append(ipv6block, &ec2.SubnetIpv6CidrBlockAssociation{
			Ipv6CidrBlock: _a0.Ipv6CidrBlock,
			AssociationId: &ipv6associationid,
			Ipv6CidrBlockState: &ec2.SubnetCidrBlockState{
				State: &defaultState,
			},
		})
	}
	subnet := &ec2.Subnet{
		CidrBlock:                   _a0.CidrBlock,
		AvailabilityZone:            _a0.AvailabilityZone,
		VpcId:                       _a0.VpcId,
		SubnetId:                    &subnetId,
		Ipv6CidrBlockAssociationSet: ipv6block,
	}

	_m.vpcassocaiatedsubnet[*_a0.VpcId] = append(_m.vpcassocaiatedsubnet[*_a0.VpcId], subnet)
	_m.subnets[subnetId] = subnet
	output = &ec2.CreateSubnetOutput{
		Subnet: subnet,
	}
	return
}

// DescribeSubnets provides a mock function with given fields: _a0
func (_m *EC2API) DescribeSubnets(_a0 *ec2.DescribeSubnetsInput) (output *ec2.DescribeSubnetsOutput, err error) {
	output = &ec2.DescribeSubnetsOutput{}
	if err := _m.recorder.CheckError("DescribeSubnets"); err != nil {
		return output, err
	}
	_m.recorder.Record("DescribeSubnets")
	returns, exist := _m.recorder.giveRecordedOutput("DescribeSubnets", _a0)
	if exist {

		assertedErr, _ := returns[1].(error)
		return returns[0].(*ec2.DescribeSubnetsOutput), assertedErr
	}
	filteredSubnets := []*ec2.Subnet{}
	for _, val := range _a0.SubnetIds {
		for _, associatedsubnet := range _m.vpcassocaiatedsubnet {
			for _, subnet := range associatedsubnet {
				if *val == *subnet.SubnetId {
					filteredSubnets = append(filteredSubnets, subnet)
				}
			}
		}
	}
	if len(_a0.Filters) != 0 {
		output.Subnets = filteredSubnets
		return
	}
	var doesHaveSupportedFilter bool
	for _, filter := range _a0.Filters {
		exist, _ := in_array(*filter.Name, supportedsubnetfilter)
		if exist {
			doesHaveSupportedFilter = true
			break
		}
		doesHaveSupportedFilter = false
	}

	if !doesHaveSupportedFilter {
		return
	}

	if len(filteredSubnets) == 0 {
		for _, associatedsubnet := range _m.vpcassocaiatedsubnet {
			for _, subnet := range associatedsubnet {
				filteredSubnets = append(filteredSubnets, subnet)
			}
		}
	}
	furtherFilteredSubnet := []*ec2.Subnet{}

	for _, val := range filteredSubnets {
		for _, filter := range _a0.Filters {
			if *filter.Name == "vpc-id" {
				for _, vpcid := range filter.Values {
					if *val.VpcId == *vpcid {
						furtherFilteredSubnet = append(furtherFilteredSubnet, val)
					}
				}
			}
		}
	}

	if len(furtherFilteredSubnet) != 0 {
		filteredSubnets = furtherFilteredSubnet
		furtherFilteredSubnet = []*ec2.Subnet{}
	}

	for _, val := range filteredSubnets {
		for _, filter := range _a0.Filters {
			if *filter.Name == "availabilityZone" {
				for _, availabilityZone := range filter.Values {
					if *val.AvailabilityZone == *availabilityZone {
						furtherFilteredSubnet = append(furtherFilteredSubnet, val)
					}
				}
			}
		}
	}

	if len(furtherFilteredSubnet) != 0 {
		filteredSubnets = furtherFilteredSubnet
		furtherFilteredSubnet = []*ec2.Subnet{}
	}

	for _, val := range filteredSubnets {
		for _, filter := range _a0.Filters {
			if *filter.Name == "cidrBlock" {
				for _, cidrBlock := range filter.Values {
					if *val.CidrBlock == *cidrBlock {
						furtherFilteredSubnet = append(furtherFilteredSubnet, val)
					}
				}
			}
		}
	}

	output.Subnets = furtherFilteredSubnet
	return
}

// CreateNetworkInterface provides a mock function with given fields: _a0
func (_m *EC2API) CreateNetworkInterface(_a0 *ec2.CreateNetworkInterfaceInput) (output *ec2.CreateNetworkInterfaceOutput, err error) {
	output = &ec2.CreateNetworkInterfaceOutput{}
	if err := _m.recorder.CheckError("CreateNetworkInterface"); err != nil {
		return output, err
	}
	_m.recorder.Record("CreateNetworkInterface")
	returns, exist := _m.recorder.giveRecordedOutput("CreateNetworkInterface", _a0)
	if exist {

		assertedErr, _ := returns[1].(error)
		return returns[0].(*ec2.CreateNetworkInterfaceOutput), assertedErr
	}
	subnet := _m.subnets[*_a0.SubnetId]
	cidr := subnet.CidrBlock
	var hostForCidr string
	// TODO: @sch00lb0y retry block is repeated many time, can be moved to a function
	// retry until you find the unassigned ip
	for {

		hostForCidr, err = PickRandomHostFromCIDR(*cidr)
		if err != nil {
			return
		}
		assignedIps, ok := _m.assignedIpOnSubnet[*_a0.SubnetId]
		if !ok {
			_m.assignedIpOnSubnet[*_a0.SubnetId] = append(_m.assignedIpOnSubnet[*_a0.SubnetId], hostForCidr)
			break
		}
		// check the random ip is alredy assiggned or not, if assingned, redo the step
		exist, _ := in_array(hostForCidr, assignedIps)
		if !exist {
			break
		}
	}
	var ntwInterfaceId string
	// retry until you find the unassinged interface id
	for {
		ntwInterfaceId = "eni-" + rangeIn(1000, 9999)
		_, exist := _m.networkinterfaces[ntwInterfaceId]
		if !exist {
			break
		}
	}

	// retry until you find the unassinged mac address
	var randomMac string
	for {
		randomMac, err = GiveRandMacAddress()
		if err != nil {
			return
		}
		exist, _ := in_array(randomMac, _m.assignedMacAddress)
		if !exist {
			_m.assignedMacAddress = append(_m.assignedMacAddress, randomMac)
			break
		}
	}
	primary := true
	privateDnsName := "ip-" + strings.Replace(hostForCidr, ".", "-", -1) + ".ap-southeast-1.compute.internal"
	ntwInterface := &ec2.NetworkInterface{
		Description:        _a0.Description,
		SubnetId:           _a0.SubnetId,
		VpcId:              subnet.VpcId,
		NetworkInterfaceId: &ntwInterfaceId,
		MacAddress:         &randomMac,
		PrivateIpAddress:   &hostForCidr,
		PrivateDnsName:     &privateDnsName,
		PrivateIpAddresses: []*ec2.NetworkInterfacePrivateIpAddress{
			&ec2.NetworkInterfacePrivateIpAddress{
				PrivateIpAddress: &hostForCidr,
				PrivateDnsName:   &privateDnsName,
				Primary:          &primary,
			},
		},
	}
	_m.networkinterfaces[ntwInterfaceId] = ntwInterface
	output.NetworkInterface = ntwInterface
	return
}

// DeleteNetworkInterface provides a mock function with given fields: _a0
func (_m *EC2API) DeleteNetworkInterface(_a0 *ec2.DeleteNetworkInterfaceInput) (output *ec2.DeleteNetworkInterfaceOutput, err error) {
	output = &ec2.DeleteNetworkInterfaceOutput{}
	if err := _m.recorder.CheckError("DeleteNetworkInterface"); err != nil {
		return output, err
	}
	_m.recorder.Record("DeleteNetworkInterface")
	returns, exist := _m.recorder.giveRecordedOutput("DeleteNetworkInterface", _a0)
	if exist {

		assertedErr, _ := returns[1].(error)
		return returns[0].(*ec2.DeleteNetworkInterfaceOutput), assertedErr
	}
	_, ok := _m.networkinterfaces[*_a0.NetworkInterfaceId]
	if !ok {
		return nil, errors.New("networks interface not found")
	}
	delete(_m.networkinterfaces, *_a0.NetworkInterfaceId)
	return
}

// AllocateAddress provides a mock function with given fields: _a0
func (_m *EC2API) AllocateAddress(_a0 *ec2.AllocateAddressInput) (output *ec2.AllocateAddressOutput, err error) {
	output = &ec2.AllocateAddressOutput{}
	if err := _m.recorder.CheckError("AllocateAddress"); err != nil {
		return output, err
	}
	_m.recorder.Record("AllocateAddress")
	returns, exist := _m.recorder.giveRecordedOutput("AllocateAddress", _a0)
	if exist {
		assertedErr, _ := returns[1].(error)
		return returns[0].(*ec2.AllocateAddressOutput), assertedErr
	}
	allocationId, err := uuid.NewV4()
	allocationIdStr := "eipalloc-" + allocationId.String()
	if err != nil {
		return
	}
	// no need to retry since uuid is unique
	var allocatedElasticIps []string
	for _, allocatedIp := range _m.assignedelasticIps {
		allocatedElasticIps = append(allocatedElasticIps, allocatedIp)
	}
	var randomElasticIp string
	// retry to find a unassinged ip
	for {
		randomElasticIp = randomdata.IpV4Address()
		exist, _ := in_array(randomElasticIp, allocatedElasticIps)
		if !exist {
			break
		}
	}
	_m.assignedelasticIps[allocationIdStr] = randomElasticIp
	output.PublicIp = &randomElasticIp
	// default vpc is used, cuz in avi only vpc is used
	// TODO: add support for standard, if needed in future
	output.Domain = &AVI_STANDARD_ELASTIC_ALLOCATION_DOMAIN
	output.AllocationId = &allocationIdStr
	return
}

// ReleaseAddress provides a mock function with given fields: _a0
func (_m *EC2API) ReleaseAddress(_a0 *ec2.ReleaseAddressInput) (output *ec2.ReleaseAddressOutput, err error) {
	output = &ec2.ReleaseAddressOutput{}
	if err := _m.recorder.CheckError("ReleaseAddress"); err != nil {
		return output, err
	}
	_m.recorder.Record("ReleaseAddress")
	returns, exist := _m.recorder.giveRecordedOutput("ReleaseAddress", _a0)
	if exist {
		assertedErr, _ := returns[1].(error)
		return returns[0].(*ec2.ReleaseAddressOutput), assertedErr
	}
	_, ok := _m.assignedelasticIps[*_a0.AllocationId]
	if !ok {
		err = errors.New("allocation id not exist")
		return
	}
	delete(_m.assignedelasticIps, *_a0.AllocationId)
	return
}

// CreateSecurityGroup provides a mock function with given fields: _a0
func (_m *EC2API) CreateSecurityGroup(_a0 *ec2.CreateSecurityGroupInput) (output *ec2.CreateSecurityGroupOutput, err error) {
	output = &ec2.CreateSecurityGroupOutput{}
	if err := _m.recorder.CheckError("CreateSecurityGroup"); err != nil {
		return output, err
	}
	_m.recorder.Record("CreateSecurityGroup")
	returns, exist := _m.recorder.giveRecordedOutput("CreateSecurityGroup", _a0)
	if exist {
		assertedErr, _ := returns[1].(error)
		return returns[0].(*ec2.CreateSecurityGroupOutput), assertedErr
	}
	securityGroupId, err := uuid.NewV4()
	securityGroupIdStr := "sg-" + securityGroupId.String()
	if err != nil {
		return
	}
	_m.assignedsecurityGroups[securityGroupIdStr] = &ec2.SecurityGroup{
		GroupId:   &securityGroupIdStr,
		VpcId:     _a0.VpcId,
		GroupName: _a0.GroupName,
	}
	output.GroupId = &securityGroupIdStr
	return
}

// DeleteSecurityGroup provides a mock function with given fields: _a0
func (_m *EC2API) DeleteSecurityGroup(_a0 *ec2.DeleteSecurityGroupInput) (output *ec2.DeleteSecurityGroupOutput, err error) {
	output = &ec2.DeleteSecurityGroupOutput{}
	if err := _m.recorder.CheckError("DeleteSecurityGroup"); err != nil {
		return output, err
	}
	_m.recorder.Record("DeleteSecurityGroup")
	returns, exist := _m.recorder.giveRecordedOutput("DeleteSecurityGroup", _a0)
	if exist {
		assertedErr, _ := returns[1].(error)
		return returns[0].(*ec2.DeleteSecurityGroupOutput), assertedErr
	}
	_, exist = _m.assignedsecurityGroups[*_a0.GroupId]
	if !exist {
		err = errors.New("group id not exist")
		return
	}
	delete(_m.assignedsecurityGroups, *_a0.GroupId)
	return
}

// DescribeSecurityGroups provides a mock function with given fields: _a0
func (_m *EC2API) DescribeSecurityGroups(_a0 *ec2.DescribeSecurityGroupsInput) (output *ec2.DescribeSecurityGroupsOutput, err error) {
	// TODO @sch00lb0y check what all the filters are being used in avi and implement according to that
	// for referenece check DescribeSubnets mock
	output = &ec2.DescribeSecurityGroupsOutput{}
	if err := _m.recorder.CheckError("DescribeSecurityGroups"); err != nil {
		return output, err
	}
	_m.recorder.Record("DescribeSecurityGroups")
	returns, exist := _m.recorder.giveRecordedOutput("DescribeSecurityGroups", _a0)
	if exist {
		assertedErr, _ := returns[1].(error)
		return returns[0].(*ec2.DescribeSecurityGroupsOutput), assertedErr
	}

	for _, groupId := range _a0.GroupIds {
		sg, ok := _m.assignedsecurityGroups[*groupId]
		if ok {
			output.SecurityGroups = append(output.SecurityGroups, sg)
		}
	}
	return
}

// AuthorizeSecurityGroupIngress provides a mock function with given fields: _a0
func (_m *EC2API) AuthorizeSecurityGroupIngress(_a0 *ec2.AuthorizeSecurityGroupIngressInput) (output *ec2.AuthorizeSecurityGroupIngressOutput, err error) {
	output = &ec2.AuthorizeSecurityGroupIngressOutput{}
	if err := _m.recorder.CheckError("AuthorizeSecurityGroupIngress"); err != nil {
		return output, err
	}
	_m.recorder.Record("AuthorizeSecurityGroupIngress")
	returns, exist := _m.recorder.giveRecordedOutput("AuthorizeSecurityGroupIngress", _a0)
	if exist {
		assertedErr, _ := returns[1].(error)
		return returns[0].(*ec2.AuthorizeSecurityGroupIngressOutput), assertedErr
	}
	securityGroup, exist := _m.assignedsecurityGroups[*_a0.GroupId]
	if !exist {
		err = errors.New("Group ID not exist")
	}

	ingressRule := &ec2.IpPermission{
		IpProtocol: _a0.IpProtocol,
		FromPort:   _a0.FromPort,
		ToPort:     _a0.ToPort,
		IpRanges: []*ec2.IpRange{
			&ec2.IpRange{
				CidrIp: _a0.CidrIp,
			},
		},
	}
	if len(securityGroup.IpPermissions) == 0 {
		securityGroup.IpPermissions = make([]*ec2.IpPermission, 0)
	}
	securityGroup.IpPermissions = append(securityGroup.IpPermissions, ingressRule)
	_m.assignedsecurityGroups[*_a0.GroupId] = securityGroup
	return
}

// RevokeSecurityGroupIngress provides a mock function with given fields: _a0
func (_m *EC2API) RevokeSecurityGroupIngress(_a0 *ec2.RevokeSecurityGroupIngressInput) (output *ec2.RevokeSecurityGroupIngressOutput, err error) {
	output = &ec2.RevokeSecurityGroupIngressOutput{}
	if err := _m.recorder.CheckError("RevokeSecurityGroupIngress"); err != nil {
		return output, err
	}
	_m.recorder.Record("RevokeSecurityGroupIngress")
	returns, exist := _m.recorder.giveRecordedOutput("RevokeSecurityGroupIngress", _a0)
	if exist {
		assertedErr, _ := returns[1].(error)
		return returns[0].(*ec2.RevokeSecurityGroupIngressOutput), assertedErr
	}
	securityGroup, exist := _m.assignedsecurityGroups[*_a0.GroupId]
	if !exist {
		err = errors.New("Group ID not exist")
	}
	var matchingForAll bool
	var foundIndex int
	for index, ingressRule := range securityGroup.IpPermissions {
		matchingForAll = false
		if ingressRule.IpProtocol != _a0.IpProtocol || ingressRule.FromPort != _a0.FromPort || ingressRule.ToPort != _a0.ToPort {
			matchingForAll = true
			foundIndex = index
			break
		}
	}
	if matchingForAll {
		securityGroup.IpPermissions = append(securityGroup.IpPermissions[:foundIndex], securityGroup.IpPermissions[foundIndex+1:]...)
	}
	_m.assignedsecurityGroups[*_a0.GroupId] = securityGroup
	return
}

// AssignPrivateIpAddresses provides a mock function with given fields: _a0
func (_m *EC2API) AssignPrivateIpAddresses(_a0 *ec2.AssignPrivateIpAddressesInput) (output *ec2.AssignPrivateIpAddressesOutput, err error) {
	output = &ec2.AssignPrivateIpAddressesOutput{}
	if err := _m.recorder.CheckError("AssignPrivateIpAddresses"); err != nil {
		return output, err
	}
	_m.recorder.Record("AssignPrivateIpAddresses")
	returns, exist := _m.recorder.giveRecordedOutput("AssignPrivateIpAddresses", _a0)
	if exist {
		assertedErr, _ := returns[1].(error)
		return returns[0].(*ec2.AssignPrivateIpAddressesOutput), assertedErr
	}
	networkInterface, exist := _m.networkinterfaces[*_a0.NetworkInterfaceId]
	if !exist {
		err = errors.New("interface id not found")
		return
	}
	primary := false
	if _a0.SecondaryPrivateIpAddressCount == nil {
		//handle for attaching ip
		for _, ip := range _a0.PrivateIpAddresses {
			if networkInterface.PrivateIpAddresses == nil {
				networkInterface.PrivateIpAddresses = make([]*ec2.NetworkInterfacePrivateIpAddress, 0)
			}
			privateDnsName := "ip-" + strings.Replace(*ip, ".", "-", -1) + ".ap-southeast-1.compute.internal"
			networkInterface.PrivateIpAddresses = append(networkInterface.PrivateIpAddresses, &ec2.NetworkInterfacePrivateIpAddress{
				PrivateIpAddress: ip,
				PrivateDnsName:   &privateDnsName,
				Primary:          &primary,
			})
		}

		_m.networkinterfaces[*_a0.NetworkInterfaceId] = networkInterface
		return
	}
	count := *_a0.SecondaryPrivateIpAddressCount

	subnet := _m.subnets[*networkInterface.SubnetId]
	cidr := subnet.CidrBlock
	randomSecondaryIps := []string{}
	var i int64
	for i = 0; i < count; i++ {
		var hostForCidr string

		for {
			hostForCidr, err = PickRandomHostFromCIDR(*cidr)
			if err != nil {
				return
			}
			assignedIps, ok := _m.assignedIpOnSubnet[*subnet.SubnetId]
			if !ok {
				_m.assignedIpOnSubnet[*subnet.SubnetId] = append(_m.assignedIpOnSubnet[*subnet.SubnetId], hostForCidr)
				break
			}
			// check the random ip is alredy assiggned or not, if assingned, redo the step
			exist, _ := in_array(hostForCidr, assignedIps)
			if !exist {
				randomSecondaryIps = append(randomSecondaryIps, hostForCidr)
				break
			}
			time.Sleep(1 * time.Second)
		}
	}
	for _, secondaryip := range randomSecondaryIps {
		privateDnsName := "ip-" + strings.Replace(secondaryip, ".", "-", -1) + ".ap-southeast-1.compute.internal"
		networkInterface.PrivateIpAddresses = append(networkInterface.PrivateIpAddresses, &ec2.NetworkInterfacePrivateIpAddress{
			PrivateIpAddress: &secondaryip,
			PrivateDnsName:   &privateDnsName,
			Primary:          &primary,
		})
	}
	_m.networkinterfaces[*_a0.NetworkInterfaceId] = networkInterface
	return
}

// UnassignPrivateIpAddresses provides a mock function with given fields: _a0
func (_m *EC2API) UnassignPrivateIpAddresses(_a0 *ec2.UnassignPrivateIpAddressesInput) (output *ec2.UnassignPrivateIpAddressesOutput, err error) {
	output = &ec2.UnassignPrivateIpAddressesOutput{}
	if err := _m.recorder.CheckError("UnassignPrivateIpAddresses"); err != nil {
		return output, err
	}
	_m.recorder.Record("UnassignPrivateIpAddresses")
	returns, exist := _m.recorder.giveRecordedOutput("UnassignPrivateIpAddresses", _a0)
	if exist {
		assertedErr, _ := returns[1].(error)
		return returns[0].(*ec2.UnassignPrivateIpAddressesOutput), assertedErr
	}
	networkInterface, exist := _m.networkinterfaces[*_a0.NetworkInterfaceId]
	if !exist {
		err = errors.New("interface id not found")
	}
	var foundIndex int
	var found bool
	for _, incommingPrivateIp := range _a0.PrivateIpAddresses {
		for index, privateIp := range networkInterface.PrivateIpAddresses {
			found = false
			if privateIp.PrivateIpAddress == incommingPrivateIp {
				foundIndex = index
				found = true
				break
			}
			if found {
				networkInterface.PrivateIpAddresses = append(networkInterface.PrivateIpAddresses[:foundIndex], networkInterface.PrivateIpAddresses[foundIndex+1:]...)
			}
		}
	}
	return
}

// DescribeInstances provides a mock function with given fields: _a0
func (_m *EC2API) DescribeInstances(_a0 *ec2.DescribeInstancesInput) (output *ec2.DescribeInstancesOutput, err error) {
	output = &ec2.DescribeInstancesOutput{}
	if err := _m.recorder.CheckError("DescribeInstances"); err != nil {
		return output, err
	}
	_m.recorder.Record("DescribeInstances")
	returns, exist := _m.recorder.giveRecordedOutput("DescribeInstances", _a0)
	if exist {
		assertedErr, _ := returns[1].(error)
		return returns[0].(*ec2.DescribeInstancesOutput), assertedErr
	}
	aggregatedInstances := []*ec2.Instance{}
	for _, instanceId := range _a0.InstanceIds {
		for _, instance := range _m.createdEc2instances {
			if *instance.InstanceId == *instanceId {
				aggregatedInstances = append(aggregatedInstances, instance)
			}
		}
	}
	reservations := []*ec2.Reservation{
		&ec2.Reservation{
			Instances: aggregatedInstances,
		},
	}
	output.Reservations = reservations
	return
}

// DescribeInstanceAttribute provides a mock function with given fields: _a0
func (_m *EC2API) DescribeInstanceAttribute(_a0 *ec2.DescribeInstanceAttributeInput) (output *ec2.DescribeInstanceAttributeOutput, err error) {
	output = &ec2.DescribeInstanceAttributeOutput{}
	if err := _m.recorder.CheckError("DescribeInstanceAttribute"); err != nil {
		return output, err
	}
	_m.recorder.Record("DescribeInstanceAttribute")
	returns, exist := _m.recorder.giveRecordedOutput("DescribeInstanceAttribute", _a0)
	if exist {
		assertedErr, _ := returns[1].(error)
		return returns[0].(*ec2.DescribeInstanceAttributeOutput), assertedErr
	}
	//NOTE: support for GroupSet only added, add remaining attribute if needed.
	var groupSet string = "groupSet"
	if _a0.Attribute == &groupSet {
		for _, instance := range _m.createdEc2instances {
			if instance.InstanceId == _a0.InstanceId {
				output.Groups = instance.SecurityGroups
			}
		}
	}
	return
}

func (_m *EC2API) CreateTags(_a0 *ec2.CreateTagsInput) (output *ec2.CreateTagsOutput, err error) {
	output = &ec2.CreateTagsOutput{}
	if err := _m.recorder.CheckError("CreateTags"); err != nil {
		return output, err
	}
	_m.recorder.Record("CreateTags")
	returns, exist := _m.recorder.giveRecordedOutput("CreateTags", _a0)
	if exist {
		assertedErr, _ := returns[1].(error)
		return returns[0].(*ec2.CreateTagsOutput), assertedErr
	}
	//if prefix is eni then update network interface tag
	for _, resourceId := range _a0.Resources {
		if strings.HasPrefix(*resourceId, "eni") {
			// update tag for network interface
			networkInterfaceCard, ok := _m.networkinterfaces[*resourceId]
			if !ok {
				return output, errors.New("network interface not found")
			}
			networkInterfaceCard.TagSet = _a0.Tags
			_m.networkinterfaces[*resourceId] = networkInterfaceCard
		}
	}
	return
}

// DescribeNetworkInterfaces provides a mock function with given fields: _a0
func (_m *EC2API) DescribeNetworkInterfaces(_a0 *ec2.DescribeNetworkInterfacesInput) (output *ec2.DescribeNetworkInterfacesOutput, err error) {
	output = &ec2.DescribeNetworkInterfacesOutput{}
	if err := _m.recorder.CheckError("DescribeNetworkInterfaces"); err != nil {
		return output, err
	}
	_m.recorder.Record("DescribeNetworkInterfaces")
	returns, exist := _m.recorder.giveRecordedOutput("DescribeNetworkInterfaces", _a0)
	if exist {
		assertedErr, _ := returns[1].(error)
		return returns[0].(*ec2.DescribeNetworkInterfacesOutput), assertedErr
	}

	filteredNetworkInterfaces := []*ec2.NetworkInterface{}
	for _, id := range _a0.NetworkInterfaceIds {
		val, ok := _m.networkinterfaces[*id]
		if ok {
			filteredNetworkInterfaces = append(filteredNetworkInterfaces, val)
		}
	}
	if len(_a0.Filters) == 0 {
		output.NetworkInterfaces = filteredNetworkInterfaces
		return
	}

	if len(filteredNetworkInterfaces) == 0 {
		for _, val := range _m.networkinterfaces {
			filteredNetworkInterfaces = append(filteredNetworkInterfaces, val)
		}
	}

	furtherFilteredNetworkInterface := []*ec2.NetworkInterface{}

	for _, val := range filteredNetworkInterfaces {
		for _, filter := range _a0.Filters {
			if *filter.Name == "vpc-id" {
				for _, vpcid := range filter.Values {
					if *val.VpcId == *vpcid {
						furtherFilteredNetworkInterface = append(furtherFilteredNetworkInterface, val)
					}
				}
			}
		}
	}

	if len(furtherFilteredNetworkInterface) != 0 {
		filteredNetworkInterfaces = furtherFilteredNetworkInterface
		furtherFilteredNetworkInterface = []*ec2.NetworkInterface{}
	}

	for _, val := range filteredNetworkInterfaces {
		for _, filter := range _a0.Filters {
			if *filter.Name == "subnet-id" {
				for _, subnetid := range filter.Values {
					if *val.SubnetId == *subnetid {
						furtherFilteredNetworkInterface = append(furtherFilteredNetworkInterface, val)
					}
				}
			}
		}
	}
	if len(furtherFilteredNetworkInterface) != 0 {
		filteredNetworkInterfaces = furtherFilteredNetworkInterface
		furtherFilteredNetworkInterface = []*ec2.NetworkInterface{}
	} else {
		furtherFilteredNetworkInterface = filteredNetworkInterfaces
	}

	for _, val := range filteredNetworkInterfaces {
		for _, filter := range _a0.Filters {
			if *filter.Name == "availabilityZone" {
				for _, availabilityZone := range filter.Values {
					if *val.AvailabilityZone == *availabilityZone {
						furtherFilteredNetworkInterface = append(furtherFilteredNetworkInterface, val)
					}
				}
			}
		}
	}

	if len(furtherFilteredNetworkInterface) != 0 {
		filteredNetworkInterfaces = furtherFilteredNetworkInterface
		furtherFilteredNetworkInterface = []*ec2.NetworkInterface{}
	} else {
		furtherFilteredNetworkInterface = filteredNetworkInterfaces
	}

	for _, val := range filteredNetworkInterfaces {
		for _, filter := range _a0.Filters {
			if *filter.Name == "availabilityZone" {
				for _, availabilityZone := range filter.Values {
					if *val.AvailabilityZone == *availabilityZone {
						furtherFilteredNetworkInterface = append(furtherFilteredNetworkInterface, val)
					}
				}
			}
		}
	}
	if len(furtherFilteredNetworkInterface) != 0 {
		filteredNetworkInterfaces = furtherFilteredNetworkInterface
		furtherFilteredNetworkInterface = []*ec2.NetworkInterface{}
	} else {
		furtherFilteredNetworkInterface = filteredNetworkInterfaces
	}

	for _, val := range filteredNetworkInterfaces {
		for _, filter := range _a0.Filters {
			if strings.HasPrefix("tag", *filter.Name) {
				// get the tag name
				tagName := strings.Split(*filter.Name, ":")[1]
				tagValue := *filter.Values[0]
				for _, tag := range val.TagSet {
					if *tag.Key == tagName {
						if *tag.Value == tagValue {
							furtherFilteredNetworkInterface = append(furtherFilteredNetworkInterface, val)
						}
					}
				}
			}
		}
	}
	if len(furtherFilteredNetworkInterface) == 0 {
		furtherFilteredNetworkInterface = filteredNetworkInterfaces
	}

	output.NetworkInterfaces = furtherFilteredNetworkInterface
	return
}

// AssociateAddress provides a mock function with given fields: _a0
func (_m *EC2API) AssociateAddress(_a0 *ec2.AssociateAddressInput) (output *ec2.AssociateAddressOutput, err error) {
	output = &ec2.AssociateAddressOutput{}
	if err := _m.recorder.CheckError("AssociateAddress"); err != nil {
		return output, err
	}
	_m.recorder.Record("AssociateAddress")
	returns, exist := _m.recorder.giveRecordedOutput("AssociateAddress", _a0)
	if exist {
		assertedErr, _ := returns[1].(error)
		return returns[0].(*ec2.AssociateAddressOutput), assertedErr
	}
	networkInterfaceCard, ok := _m.networkinterfaces[*_a0.NetworkInterfaceId]
	if !ok {
		return output, errors.New("network interface not found")
	}
	for i, privateIP := range networkInterfaceCard.PrivateIpAddresses {
		if *privateIP.PrivateIpAddress == *_a0.PrivateIpAddress {
			association := &ec2.NetworkInterfaceAssociation{}
			association.AllocationId = _a0.AllocationId
			association.PublicIp = _a0.PublicIp
			associationId, err := uuid.NewV4()
			if err != nil {
				return output, err
			}
			associationIdStr := "fip-alloc-" + associationId.String()
			association.AssociationId = &associationIdStr
			privateIP.Association = association
			networkInterfaceCard.PrivateIpAddresses[i] = privateIP
		}
	}
	return
}

// DescribeAddresses provides a mock function with given fields: _a0
func (_m *EC2API) DescribeAddresses(_a0 *ec2.DescribeAddressesInput) (output *ec2.DescribeAddressesOutput, err error) {
	output = &ec2.DescribeAddressesOutput{}
	output.Addresses = make([]*ec2.Address, 0)
	if err := _m.recorder.CheckError("DescribeAddresses"); err != nil {
		return output, err
	}
	_m.recorder.Record("DescribeAddresses")
	returns, exist := _m.recorder.giveRecordedOutput("DescribeAddresses", _a0)
	if exist {
		assertedErr, _ := returns[1].(error)
		return returns[0].(*ec2.DescribeAddressesOutput), assertedErr
	}
	for _, inputIP := range _a0.PublicIps {
		for allocationId, elasticIP := range _m.assignedelasticIps {
			if elasticIP == *inputIP {
				output.Addresses = append(output.Addresses, &ec2.Address{
					AllocationId: &allocationId,
					PublicIp:     inputIP,
				})
			}
		}
	}
	return
}

// DisassociateAddress provides a mock function with given fields: _a0
func (_m *EC2API) DisassociateAddress(_a0 *ec2.DisassociateAddressInput) (output *ec2.DisassociateAddressOutput, err error) {
	output = &ec2.DisassociateAddressOutput{}
	if err := _m.recorder.CheckError("AssociateAddress"); err != nil {
		return output, err
	}
	_m.recorder.Record("AssociateAddress")
	returns, exist := _m.recorder.giveRecordedOutput("AssociateAddress", _a0)
	if exist {
		assertedErr, _ := returns[1].(error)
		return returns[0].(*ec2.DisassociateAddressOutput), assertedErr
	}
	for networkInterfaceIndex, networkInterfaceCard := range _m.networkinterfaces {
		for i, privateIP := range networkInterfaceCard.PrivateIpAddresses {
			if *privateIP.Association.AllocationId == *_a0.AssociationId {
				networkInterfaceCard.PrivateIpAddresses[i].Association = &ec2.NetworkInterfaceAssociation{}
				_m.networkinterfaces[networkInterfaceIndex] = networkInterfaceCard
			}
		}
	}
	return
}
