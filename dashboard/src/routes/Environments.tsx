import { TableColumn } from "react-data-table-component";
import ResourceView from "../components/ResourceView";
import { EnvironmentProperties } from "../resources/environments";
import { extractResourceGroup, Resource } from "../resources/resources";

const columns : TableColumn<Resource<EnvironmentProperties>>[] = [
  {
      name: 'Name',
      selector: (row: Resource<EnvironmentProperties>) => row.name,
      sortable: true,
  },
  {
      name: 'Resource Group',
      selector: (row: Resource<EnvironmentProperties>) => extractResourceGroup(row.id) ?? '',
      sortable: true,
  },
];


export default function EnvironmentPage() {
  return (
    <>
      <ResourceView 
        columns={columns} 
        heading="Environments"
        resourceType="Applications.Core/environments" 
        selectionMessage="Select an environment to display details..."/>
    </>
  );
}