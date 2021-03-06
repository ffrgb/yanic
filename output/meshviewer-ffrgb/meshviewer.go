package meshviewerFFRGB

import (
	"fmt"
	"strings"

	"github.com/bdlm/log"

	"github.com/FreifunkBremen/yanic/lib/jsontime"
	"github.com/FreifunkBremen/yanic/runtime"
)

const (
	LINK_TYPE_WIRELESS = "wifi"
	LINK_TYPE_TUNNEL   = "vpn"
	LINK_TYPE_FALLBACK = "other"
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
						typeList[addr] = LINK_TYPE_WIRELESS
					}
					for _, addr := range mesh.Interfaces.Tunnel {
						typeList[addr] = LINK_TYPE_TUNNEL
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

			if link := links[key]; link != nil {
				linkType, linkTypeFound := typeList[linkOrigin.SourceAddress]
				if !linkTypeFound {
					linkType, linkTypeFound = typeList[linkOrigin.TargetAddress]
				}

				if switchSourceTarget {
					link.TargetTQ = linkOrigin.TQ

					linkType, linkTypeFound = typeList[linkOrigin.TargetAddress]
					if !linkTypeFound {
						linkType, linkTypeFound = typeList[linkOrigin.SourceAddress]
					}
				} else {
					link.SourceTQ = linkOrigin.TQ
				}

				if linkTypeFound && linkType != link.Type {
					if link.Type == LINK_TYPE_FALLBACK {
						link.Type = linkType
					} else {
						log.WithFields(map[string]interface{}{
							"link":   fmt.Sprintf("%s-%s", linkOrigin.SourceAddress, linkOrigin.TargetAddress),
							"prev":   link.Type,
							"new":    linkType,
							"source": typeList[linkOrigin.SourceAddress],
							"target": typeList[linkOrigin.TargetAddress],
						}).Warn("different linktypes")
					}
				}

				continue
			}
			link := &Link{
				Source:        linkOrigin.SourceID,
				SourceAddress: linkOrigin.SourceAddress,
				Target:        linkOrigin.TargetID,
				TargetAddress: linkOrigin.TargetAddress,
				SourceTQ:      linkOrigin.TQ,
				TargetTQ:      0,
			}

			linkType, linkTypeFound := typeList[linkOrigin.SourceAddress]
			if !linkTypeFound {
				linkType, linkTypeFound = typeList[linkOrigin.TargetAddress]
			}

			if switchSourceTarget {
				link.SourceTQ = 0
				link.Source = linkOrigin.TargetID
				link.SourceAddress = linkOrigin.TargetAddress
				link.TargetTQ = linkOrigin.TQ
				link.Target = linkOrigin.SourceID
				link.TargetAddress = linkOrigin.SourceAddress

				linkType, linkTypeFound = typeList[linkOrigin.TargetAddress]
				if !linkTypeFound {
					linkType, linkTypeFound = typeList[linkOrigin.SourceAddress]
				}
			}

			if linkTypeFound {
				link.Type = linkType
			} else {
				link.Type = LINK_TYPE_FALLBACK
			}
			links[key] = link
			meshviewer.Links = append(meshviewer.Links, link)
		}
	}

	return meshviewer
}
