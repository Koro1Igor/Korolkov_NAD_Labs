import { Nav, Navbar } from 'react-bootstrap'
import { NavLink } from 'react-router-dom'

export function BottomNavbar() {
  return (
    <Navbar fixed="bottom" className="bottom-nav border-top bg-white">
      <Nav className="w-100 d-flex justify-content-around">
        <Nav.Link as={NavLink} to="/feed">Лента</Nav.Link>
        <Nav.Link disabled>Добавление</Nav.Link>
        <Nav.Link as={NavLink} to="/launch-vehicles">Плитка</Nav.Link>
        <Nav.Link disabled>Вход</Nav.Link>
      </Nav>
    </Navbar>
  )
}
