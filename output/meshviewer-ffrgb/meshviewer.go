package meshviewerFFRGB

import (
	"fmt"
	"log"
	"strings"

	"github.com/FreifunkBremen/yanic/lib/jsontime"
	"github.com/FreifunkBremen/yanic/runtime"
)

func transform(nodes *runtime.Nodes) *Meshviewer {

	meshviewer := &Meshviewer{
		Timestamp: jsontime.Now(),
		Nodes:     make([]*Node, 0),
		Links:     make([]*Link, 0),
	}

	links := make(map[string]*Link)
	typeList := make(map[string]string)

	nodes.RLock()
	defer nodes.RUnlock()

	for _, nodeOrigin := range nodes.List {
		node := NewNode(nodes, nodeOrigin)
		meshviewer.Nodes = append(meshviewer.Nodes, node)

		if !nodeOrigin.Online {
			continue
		}

		if nodeinfo := nodeOrigin.Nodeinfo; nodeinfo != nil {
			if meshes := nodeinfo.Network.Mesh; meshes != nil {
				for _, mesh := range meshes {
					for _, addr := range mesh.Interfaces.Wireless {
						typeList[addr] = "wifi"
					}
					for _, addr := range mesh.Interfaces.Tunnel {
						typeList[addr] = "vpn"
					}
				}
			}
		}

		for _, linkOrigin := range nodes.NodeLinks(nodeOrigin) {
			var key string
			// keep source and target in the same order
			switchSourceTarget := strings.Compare(linkOrigin.SourceAddress, linkOrigin.TargetAddress) > 0
			if switchSourceTarget {
				key = fmt.Sprintf("%s-%s", linkOrigin.SourceAddress, linkOrigin.TargetAddress)
			} else {
				key = fmt.Sprintf("%s-%s", linkOrigin.TargetAddress, linkOrigin.SourceAddress)
			}

			tq := float32(linkOrigin.TQ)

			if link := links[key]; link != nil {
				if switchSourceTarget {
					link.TargetTQ = tq
					if link.Type == "other" {
						link.Type = typeList[linkOrigin.TargetAddress]
					} else if link.Type != typeList[linkOrigin.TargetAddress] {
						log.Printf("different linktypes %s:%s current: %s source: %s target: %s", linkOrigin.SourceAddress, linkOrigin.TargetAddress, link.Type, typeList[linkOrigin.SourceAddress], typeList[linkOrigin.TargetAddress])
					}
				} else {
					link.SourceTQ = tq
					if link.Type == "other" {
						link.Type = typeList[linkOrigin.SourceAddress]
					} else if link.Type != typeList[linkOrigin.SourceAddress] {
						log.Printf("different linktypes %s:%s current: %s source: %s target: %s", linkOrigin.SourceAddress, linkOrigin.TargetAddress, link.Type, typeList[linkOrigin.SourceAddress], typeList[linkOrigin.TargetAddress])
					}
				}
				if link.Type == "" {
					link.Type = "other"
				}
				continue
			}
			link := &Link{
				Type:          typeList[linkOrigin.SourceAddress],
				Source:        linkOrigin.SourceID,
				SourceAddress: linkOrigin.SourceAddress,
				Target:        linkOrigin.TargetID,
				TargetAddress: linkOrigin.TargetAddress,
				SourceTQ:      linkOrigin.TQ,
				TargetTQ:      linkOrigin.TQ,
			}
			if switchSourceTarget {
				link.Type = typeList[linkOrigin.TargetAddress]
				link.Source = linkOrigin.TargetID
				link.SourceAddress = linkOrigin.TargetAddress
				link.Target = linkOrigin.SourceID
				link.TargetAddress = linkOrigin.SourceAddress
			}
			if link.Type == "" {
				link.Type = "other"
			}
			links[key] = link
			meshviewer.Links = append(meshviewer.Links, link)
		}
	}

	return meshviewer
}
