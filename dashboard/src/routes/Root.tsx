import { Link, NavLink, Outlet, useNavigation } from "react-router-dom";
import Container from 'react-bootstrap/Container';
import Navbar from "react-bootstrap/Navbar";

import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faHouse, faCube, faPager, faCubes, faBorderAll } from '@fortawesome/free-solid-svg-icons';

import 'bootstrap/dist/css/bootstrap.min.css';
import './Root.css';

export default function Root() {
  const navigation = useNavigation();
  return (
    <>
      <Navbar bg="primary" variant="dark">
        <Container fluid>
          <Navbar.Brand href="/">Project Radius</Navbar.Brand>
        </Container>
      </Navbar>
      <div className="Root-container">
        <div className="Root-sidebar">
          <h1>Radius Dashboard</h1>
          <nav>
            <ul>
              <li>
                <NavLink to={`environments`}><FontAwesomeIcon icon={faCube} />  Environments</NavLink>
              </li>
              <li>
                <NavLink to={`applications`}><FontAwesomeIcon icon={faPager} />  Applications</NavLink>
              </li>
              <li>
                <NavLink to={`containers`}><FontAwesomeIcon icon={faCubes} />  Containers</NavLink>
              </li>
              <li>
                <NavLink to={`resources`}><FontAwesomeIcon icon={faBorderAll} />  Resources</NavLink>
              </li>
            </ul>
          </nav>
        </div>
        <div className={navigation.state === "loading" ? "Root-detail loading" : "Root-detail"}>
          <Outlet />
        </div>
      </div>
    </>
  );
}