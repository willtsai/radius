import { Link, NavLink, Outlet, useNavigation } from "react-router-dom";
import './Root.css';

export default function Root() {
  const navigation = useNavigation();
  return (
    <>
      <div className="Root-container">
        <div className="Root-sidebar">
          <h1>Radius Dashboard</h1>
          <nav>
            <ul>
              <li>
                <Link to="/" >Home</Link>
              </li>
              <li>
                <NavLink to={`environments`}>Environments</NavLink>
              </li>
              <li>
                <NavLink to={`applications`}>Applications</NavLink>
              </li>
              <li>
                <NavLink to={`containers`}>Containers</NavLink>
              </li>
              <li>
                <NavLink to={`resources`}>Resources</NavLink>
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