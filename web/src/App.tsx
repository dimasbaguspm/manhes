import { BrowserRouter, Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import LibraryPage from './pages/LibraryPage'
import DiscoverPage from './pages/DiscoverPage'
import MangaPage from './pages/MangaPage'
import ReaderPage from './pages/ReaderPage'

export default function App() {
  return (
    <BrowserRouter future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<LibraryPage />} />
          <Route path="/discover" element={<DiscoverPage />} />
          <Route path="/manga/:mangaId" element={<MangaPage />} />
        </Route>
        <Route path="/read/:chapterId" element={<ReaderPage />} />
      </Routes>
    </BrowserRouter>
  )
}
