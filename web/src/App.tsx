import { BrowserRouter, Routes, Route } from 'react-router-dom'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={
          <div className="min-h-screen flex items-center justify-center bg-gray-50">
            <div className="text-center">
              <h1 className="text-4xl font-bold text-gray-900 mb-4">TempleATS</h1>
              <p className="text-gray-600">Applicant Tracking System</p>
            </div>
          </div>
        } />
      </Routes>
    </BrowserRouter>
  )
}

export default App
