import { BrowserRouter, Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import LibraryPage from './pages/library-page/library-page'
import DiscoverPage from './pages/discover-page/discover-page'
import MangaPage from './pages/manga-page/manga-page'
import ReaderPage from './pages/reader-page/reader-page'

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
