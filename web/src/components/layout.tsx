import { NavLink, Outlet } from 'react-router-dom'

export default function Layout() {
  return (
    <div className="min-h-screen bg-gray-950 text-gray-100">
      <nav className="sticky top-0 z-10 border-b border-gray-800 bg-gray-950/90 backdrop-blur">
        <div className="mx-auto flex max-w-7xl items-center gap-6 px-4 py-3">
          <span className="text-lg font-bold tracking-tight text-indigo-400">manhes</span>
          <NavLink
            to="/"
            end
            className={({ isActive }) =>
              `text-sm transition ${isActive ? 'text-white' : 'text-gray-400 hover:text-white'}`
            }
          >
            Library
          </NavLink>
          <NavLink
            to="/discover"
            className={({ isActive }) =>
              `text-sm transition ${isActive ? 'text-white' : 'text-gray-400 hover:text-white'}`
            }
          >
            Discover
          </NavLink>
        </div>
      </nav>
      <main className="mx-auto max-w-7xl px-4 py-6">
        <Outlet />
      </main>
    </div>
  )
}
