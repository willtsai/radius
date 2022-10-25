import React from "react";
import { Graph, GraphConfiguration, GraphNode, GraphLink, GraphData } from "react-d3-graph";
import { Resource } from "../resources/resources";
import "./Topology.css"

export interface ResourceNode extends GraphNode {
  resource: Resource<any>
}

export interface ResourceLink extends GraphLink {

}

const config = (size: { width: number, height: number }): Partial<GraphConfiguration<ResourceNode, ResourceLink>> => {
  return {
    // This is helpful for performance since we don't change the graph after rendering.
    directed: true,

    // This library will not behave correctly unless we set an explicit size. 
    //
    // Without a size, all of the graph nodes will be positioned in the top left corner.
    height: size.height,
    width: size.width,

    minZoom: 1,
    maxZoom: 8,
    focusZoom: 1,

    d3: {
      alphaTarget: 0.05,
      gravity: -3000,
      linkLength: 1,
      linkStrength: 1,
    },

    node: {
      renderLabel: false,
      mouseCursor: "grab",
      size: {
        width: 3000,
        height: 2000,
      },
      viewGenerator: (node: ResourceNode) => {
        return <>
          <div className="card" style={{ height: "100%", opacity: 0.8 }}>
            <div className="card-body">
              <h5 className="card-title">{node.resource.name}</h5>
              <h6 className="card-subtitle mb-2 text-muted">{node.resource.type}</h6>
              <p>This resource is really very cool.</p>
            </div>
          </div>
        </>
      },
    },
    link: {
      color: "#000000",
      //type: "CURVE_SMOOTH",
    },
  };
}

export default function TopologyView(props: GraphData<ResourceNode, ResourceLink>) {
  // The graph library needs to work with an explicit size. To handle this we wrap the graph
  // in a container div, and then size the graph to fit.
  let containerRef = React.useRef<HTMLDivElement>(null);
  let [size, setSize] = React.useState({ width: 800, height: 600 })

  React.useLayoutEffect(() => {
    if (containerRef.current) {
      setSize({ width: 800, height: 600 });
    }
  }, []);

  return <>
    <div className="TopologyView-container" ref={containerRef}>
      <Graph
        id="topology"
        data={props}
        config={config(size)} />
    </div>
  </>
}