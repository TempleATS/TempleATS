import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthContext, useAuth, useAuthProvider } from './hooks/use-auth'
import Login from './pages/Login'
import Signup from './pages/Signup'
import Dashboard from './pages/Dashboard'

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
      <Route path="/dashboard" element={
        <ProtectedRoute><Dashboard /></ProtectedRoute>
      } />
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
