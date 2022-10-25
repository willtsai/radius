import { TableColumn } from "react-data-table-component";
import ResourceView from "../components/ResourceView";
import { ApplicationProperties } from "../resources/applications";
import { extractResourceGroup, extractResourceName, Resource } from "../resources/resources";

const columns : TableColumn<Resource<ApplicationProperties>>[] = [
  {
      name: 'Name',
      selector: (row: Resource<ApplicationProperties>) => row.name,
      sortable: true,
  },
  {
      name: 'Resource Group',
      selector: (row: Resource<ApplicationProperties>) => extractResourceGroup(row.id) ?? '',
      sortable: true,
  },
  {
    name: 'Environment',
    selector: (row: Resource<ApplicationProperties>) => extractResourceName(row.properties.environment) ?? '',
    sortable: true,
  },
  {
    name:'Actions',
    cell: (row: Resource<ApplicationProperties>) => <><a className="btn btn-primary" href={`/topology?application=${encodeURIComponent(row.id)}`} role="button">Map</a></>,
    button: true,
    center: true,
  },
];


export default function ApplicationPage() {
  return (
    <>
      <ResourceView 
        columns={columns} 
        heading="Applications"
        resourceType="Applications.Core/applications" 
        selectionMessage="Select an application to display details..."/>
    </>
  );
}