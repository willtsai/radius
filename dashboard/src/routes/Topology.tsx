import React from "react";
import { useErrorHandler } from "react-error-boundary";
import { useSearchParams } from "react-router-dom";
import RefreshButton from "../components/RefreshButton";
import TopologyView, { ResourceLink, ResourceNode } from "../components/Topology";
import { isApplicationType } from "../resources/applications";
import { ContainerProperties, isContainerType } from "../resources/containers";
import { isEnvironmentType } from "../resources/environments";
import { extractResourceName, Resource } from "../resources/resources";
import "./Topology.scss"

export default function TopologyPage() {
  const [searchParams, _] = useSearchParams();
  const applicationId = searchParams.get("application")

  let [resources, setResources] = React.useState<Resource<any>[]>([]);
  let [loading, setLoading] = React.useState(true)
  let [reloadCount, setReloadCount] = React.useState(0)

  const errorHandler = useErrorHandler();

  React.useEffect(() => {
    let mounted = true;
    const fetchResources = async () => {
      setLoading(true);

      const url = '/api/resources';
      const response = await fetch(url)
      const data = await response.json();

      if (mounted) {
        setLoading(false);
        setResources(data.values);
      }
    }

    fetchResources().catch(errorHandler);
    return () => {
      mounted = false;
    }
  }, [reloadCount, errorHandler])

  const onRefreshClicked = () => {
    if (loading) {
      return
    }

    setReloadCount(reloadCount + 1);
  }

  const nodes: ResourceNode[] = [];
  resources
    .filter((value) => {
      // TODO verify resource belongs to application.
      return !isApplicationType(value) && !isEnvironmentType(value);
    })
    .forEach((value) => {
      nodes.push({ id: value.id, resource: value });
    });

  const links: ResourceLink[] = [];
  resources
    .filter((value) => {
      return isContainerType(value);
    })
    .map(value => value as Resource<ContainerProperties>)
    .forEach((value) => {
      if (!value.properties.connections) {
        return;
      }
      Object.entries(value.properties.connections).forEach(([name, connection]) => {
        links.push({ source: value.id, target: connection.source });
      });
    })

  const focusedId = nodes.length > 0 ? nodes[0].id : undefined;

  return (
    <>
      <div className="TopologyPage-header">
        <h1>{extractResourceName(applicationId!)} - Toplogy View</h1>
        <RefreshButton loading={loading} onRefreshClicked={onRefreshClicked} />
      </div>
      <div className="TopologyPage-main" >
        {(loading && resources.length === 0) ?
          <p className="TopologyPage-not-loaded">Loading resources. Please wait...</p> :
            <TopologyView nodes={nodes} links={links} focusedNodeId={focusedId}/>}
      </div>
    </>
  );
}