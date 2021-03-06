package clab

import (
	"strings"

	"github.com/awalterschulze/gographviz"
)

var g *gographviz.Graph

// GenerateGraph generates a graph for the lab topology
func (c *cLab) GenerateGraph(topo string) error {
	g = gographviz.NewGraph()
	if err := g.SetName(c.FileInfo.shortname); err != nil {
		return err
	}
	if err := g.SetDir(false); err != nil {
		return err
	}

	var attr map[string]string
	attr = make(map[string]string)
	attr["color"] = "red"
	attr["style"] = "filled"
	attr["fillcolor"] = "red"

	for nodeName, node := range c.Nodes {
		attr["label"] = nodeName
		attr["xlabel"] = node.OS
		attr["group"] = node.Group

		if strings.Contains(node.OS, "srl") {
			attr["fillcolor"] = "green"
		}
		if strings.Contains(node.Group, "bb") {
			attr["fillcolor"] = "blue"
		}
		if err := g.AddNode(c.FileInfo.shortname, node.ShortName, attr); err != nil {
			return err
		}

	}

	attr = make(map[string]string)
	attr["color"] = "green"

	for _, link := range c.Links {
		if strings.Contains(link.b.Node.ShortName, "client") {
			attr["color"] = "blue"
		}
		if err := g.AddEdge(link.a.Node.ShortName, link.b.Node.ShortName, false, attr); err != nil {
			return err
		}

	}

	// create graph directory
	CreateDirectory(c.Dir.LabGraph, 0755)

	// create graph filename
	file := c.Dir.LabGraph + "/" + c.FileInfo.name + ".dot"

	createFile(file, g.String())

	return nil
}
