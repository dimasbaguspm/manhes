import { BrowserRouter, Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import LibraryPage from './pages/LibraryPage'
import DiscoverPage from './pages/DiscoverPage'
import MangaPage from './pages/MangaPage'
import ChaptersPage from './pages/ChaptersPage'
import ReaderPage from './pages/ReaderPage'
import { MangaPagedListProvider } from './providers/MangaPagedListProvider'
import { DictionaryListProvider } from './providers/DictionaryListProvider'
import { MangaDetailProvider } from './providers/MangaDetailProvider'
import { MangaChaptersProvider } from './providers/MangaChaptersProvider'
import { MangaReaderDataProvider } from './providers/MangaReaderDataProvider'

export default function App() {
  return (
    <BrowserRouter future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={
            <MangaPagedListProvider>
              <LibraryPage />
            </MangaPagedListProvider>
          } />
          <Route path="/discover" element={
            <DictionaryListProvider>
              <DiscoverPage />
            </DictionaryListProvider>
          } />
          <Route path="/manga/:mangaId" element={
            <MangaDetailProvider>
              <MangaPage />
            </MangaDetailProvider>
          } />
          <Route path="/manga/:mangaId/:lang" element={
            <MangaChaptersProvider>
              <ChaptersPage />
            </MangaChaptersProvider>
          } />
        </Route>
        <Route path="/manga/:mangaId/:lang/read" element={
          <MangaReaderDataProvider>
            <ReaderPage />
          </MangaReaderDataProvider>
        } />
      </Routes>
    </BrowserRouter>
  )
}
