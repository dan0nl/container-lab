package clab

import (
	"os/exec"
	"strconv"

	log "github.com/sirupsen/logrus"
)

// CreateVirtualWiring provides the virtual topology between the containers
func (c *cLab) CreateVirtualWiring(id int, link *Link) (err error) {
	log.Info("Create virtual wire :", link.a.Node.ShortName, link.b.Node.ShortName, link.a.EndpointName, link.b.EndpointName)

	CreateDirectory("/run/netns/", 0755)

	var src, dst string
	var cmd *exec.Cmd

	log.Debug("Create link to /run/netns/ ", link.a.Node.LongName)
	src = "/proc/" + strconv.Itoa(link.a.Node.Pid) + "/ns/net"
	dst = "/run/netns/" + link.a.Node.LongName
	//err = linkFile(src, dst)
	cmd = exec.Command("sudo", "ln", "-s", src, dst)
	err = runCmd(cmd)
	if err != nil {
		log.Debugf("%s failed with: %v", cmd.String(), err)
	}

	log.Debug("Create link to /run/netns/ ", link.b.Node.LongName)
	src = "/proc/" + strconv.Itoa(link.b.Node.Pid) + "/ns/net"
	dst = "/run/netns/" + link.b.Node.LongName
	//err = linkFile(src, dst)
	cmd = exec.Command("sudo", "ln", "-s", src, dst)
	err = runCmd(cmd)
	if err != nil {
		log.Debugf("%s failed with: %v", cmd.String(), err)
	}

	log.Debug("create dummy veth pair")
	cmd = exec.Command("sudo", "ip", "link", "add", "dummyA", "type", "veth", "peer", "name", "dummyB")
	err = runCmd(cmd)
	if err != nil {
		log.Fatalf("%s failed with: %v", cmd.String(), err)
	}

	log.Debug("map dummy interface on container A to NS")
	cmd = exec.Command("sudo", "ip", "link", "set", "dummyA", "netns", link.a.Node.LongName)
	err = runCmd(cmd)
	if err != nil {
		log.Fatalf("%s failed with: %v", cmd.String(), err)
	}

	log.Debug("map dummy interface on container B to NS")
	cmd = exec.Command("sudo", "ip", "link", "set", "dummyB", "netns", link.b.Node.LongName)
	err = runCmd(cmd)
	if err != nil {
		log.Fatalf("%s failed with: %v", cmd.String(), err)
	}

	log.Debug("rename interface container NS A")
	cmd = exec.Command("sudo", "ip", "netns", "exec", link.a.Node.LongName, "ip", "link", "set", "dummyA", "name", link.a.EndpointName)
	err = runCmd(cmd)
	if err != nil {
		log.Fatalf("%s failed with: %v", cmd.String(), err)
	}

	log.Debug("rename interface container NS B")
	cmd = exec.Command("sudo", "ip", "netns", "exec", link.b.Node.LongName, "ip", "link", "set", "dummyB", "name", link.b.EndpointName)
	err = runCmd(cmd)
	if err != nil {
		log.Fatalf("%s failed with: %v", cmd.String(), err)
	}

	log.Debug("set interface up in container NS A")
	cmd = exec.Command("sudo", "ip", "netns", "exec", link.a.Node.LongName, "ip", "link", "set", link.a.EndpointName, "up")
	err = runCmd(cmd)
	if err != nil {
		log.Fatalf("%s failed with: %v", cmd.String(), err)
	}

	log.Debug("set interface up in container NS B")
	cmd = exec.Command("sudo", "ip", "netns", "exec", link.b.Node.LongName, "ip", "link", "set", link.b.EndpointName, "up")
	err = runCmd(cmd)
	if err != nil {
		log.Fatalf("%s failed with: %v", cmd.String(), err)
	}

	log.Debug("set RX, TX offload off on container A")
	cmd = exec.Command("docker", "exec", link.a.Node.LongName, "ethtool", "--offload", link.a.EndpointName, "rx", "off", "tx", "off")
	err = runCmd(cmd)
	if err != nil {
		log.Debugf("%s failed with: %v", cmd.String(), err)
	}

	log.Debug("set RX, TX offload off on container B")
	cmd = exec.Command("docker", "exec", link.b.Node.LongName, "ethtool", "--offload", link.b.EndpointName, "rx", "off", "tx", "off")
	err = runCmd(cmd)
	if err != nil {
		log.Debugf("%s failed with: %v", cmd.String(), err)
	}

	//ip link add tmp_a type veth peer name tmp_b
	//ip link set tmp_a netns $srl_a
	//ip link set tmp_b netns $srl_b
	//ip netns exec $srl_a ip link set tmp_a name $srl_a_int
	//ip netns exec $srl_b ip link set tmp_b name $srl_b_int
	//ip netns exec $srl_a ip link set $srl_a_int up
	//ip netns exec $srl_b ip link set $srl_b_int up
	//docker exec -ti $srl_a ethtool --offload $srl_a_int rx off tx off
	//docker exec -ti $srl_b ethtool --offload $srl_b_int rx off tx off

	return nil

}

// DeleteVirtualWiring deletes the virtual wiring
func (c *cLab) DeleteVirtualWiring(id int, link *Link) (err error) {
	log.Info("Delete virtual wire :", link.a.Node.ShortName, link.b.Node.ShortName, link.a.EndpointName, link.b.EndpointName)

	var cmd *exec.Cmd

	log.Debug("Delete netns: ", link.a.Node.LongName)
	cmd = exec.Command("sudo", "ip", "netns", "del", link.a.Node.LongName)
	err = runCmd(cmd)
	if err != nil {
		log.Debugf("%s failed with: %v", cmd.String(), err)
	}

	log.Debug("Delete netns: ", link.b.Node.LongName)
	cmd = exec.Command("sudo", "ip", "netns", "del", link.b.Node.LongName)
	err = runCmd(cmd)
	if err != nil {
		log.Debugf("%s failed with: %v", cmd.String(), err)
	}
	return nil
}

func runCmd(cmd *exec.Cmd) error {
	b, err := cmd.CombinedOutput()
	if err != nil {
		log.Debugf("'%s' failed with: %v", cmd.String(), err)
		log.Debugf("'%s' failed output: %v", cmd.String(), string(b))
		return err
	}
	log.Debugf("'%s' output: %v", cmd.String(), string(b))
	return nil
}
