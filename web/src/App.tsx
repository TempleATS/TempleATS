import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthContext, useAuth, useAuthProvider } from './hooks/use-auth'
import Login from './pages/Login'
import Signup from './pages/Signup'
import Dashboard from './pages/Dashboard'
import Reqs from './pages/Reqs'
import ReqForm from './pages/ReqForm'
import ReqDetail from './pages/ReqDetail'
import Jobs from './pages/Jobs'
import JobForm from './pages/JobForm'
import JobDetail from './pages/JobDetail'
import CareersPage from './pages/careers/CareersPage'
import CareerJobDetail from './pages/careers/CareerJobDetail'
import ApplyForm from './pages/careers/ApplyForm'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <p className="text-gray-500">Loading...</p>
      </div>
    );
  }

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

function AppRoutes() {
  const { user, loading } = useAuth();

  return (
    <Routes>
      <Route path="/login" element={
        !loading && user ? <Navigate to="/dashboard" replace /> : <Login />
      } />
      <Route path="/signup" element={
        !loading && user ? <Navigate to="/dashboard" replace /> : <Signup />
      } />

      {/* Protected routes */}
      <Route path="/dashboard" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
      <Route path="/reqs" element={<ProtectedRoute><Reqs /></ProtectedRoute>} />
      <Route path="/reqs/new" element={<ProtectedRoute><ReqForm /></ProtectedRoute>} />
      <Route path="/reqs/:reqId" element={<ProtectedRoute><ReqDetail /></ProtectedRoute>} />
      <Route path="/jobs" element={<ProtectedRoute><Jobs /></ProtectedRoute>} />
      <Route path="/jobs/new" element={<ProtectedRoute><JobForm /></ProtectedRoute>} />
      <Route path="/jobs/:jobId" element={<ProtectedRoute><JobDetail /></ProtectedRoute>} />

      {/* Public careers routes (no auth) */}
      <Route path="/careers/:orgSlug" element={<CareersPage />} />
      <Route path="/careers/:orgSlug/jobs/:jobId" element={<CareerJobDetail />} />
      <Route path="/careers/:orgSlug/jobs/:jobId/apply" element={<ApplyForm />} />

      <Route path="/" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  );
}

function App() {
  const auth = useAuthProvider();

  return (
    <AuthContext.Provider value={auth}>
      <BrowserRouter>
        <AppRoutes />
      </BrowserRouter>
    </AuthContext.Provider>
  )
}

export default App
