import { Navigate, Route, Routes } from 'react-router-dom'
import { BottomNavbar } from './components/BottomNavbar'
import { AddPage } from './pages/AddPage'
import { FeedPage } from './pages/FeedPage'
import { TilePage } from './pages/TilePage'

export default function App() {
  return (
    <>
      <Routes>
        <Route path="/" element={<Navigate to="/feed" replace />} />
        <Route path="/feed" element={<FeedPage />} />
        <Route path="/launch-vehicles" element={<TilePage />} />
        <Route path="/add" element={<AddPage />} />
      </Routes>
      <BottomNavbar />
    </>
  )
}
