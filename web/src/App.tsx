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
import JobPipeline from './pages/JobPipeline'
import RejectedList from './pages/RejectedList'
import Candidates from './pages/Candidates'
import CandidateDetail from './pages/CandidateDetail'
import ApplicationDetailPage from './pages/ApplicationDetail'
import ReqReport from './pages/ReqReport'
import Settings from './pages/Settings'
import Team from './pages/Team'
import AcceptInvite from './pages/AcceptInvite'
import Account from './pages/Account'
import Referrals from './pages/Referrals'
import Interviews from './pages/Interviews'
import ScheduleBooking from './pages/ScheduleBooking'
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

function DefaultRedirect() {
  const { user, isAtLeast } = useAuth();
  const dest = user && isAtLeast('admin') ? '/dashboard'
    : user && isAtLeast('recruiter') ? '/jobs'
    : user && isAtLeast('hiring_manager') ? '/jobs'
    : '/interviews';
  return <Navigate to={dest} replace />;
}

function AppRoutes() {
  const { user, loading } = useAuth();

  return (
    <Routes>
      <Route path="/login" element={
        !loading && user ? <DefaultRedirect /> : <Login />
      } />
      <Route path="/signup" element={
        !loading && user ? <DefaultRedirect /> : <Signup />
      } />
      <Route path="/accept-invite/:token" element={<AcceptInvite />} />

      {/* Protected routes */}
      <Route path="/dashboard" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
      <Route path="/reqs" element={<ProtectedRoute><Reqs /></ProtectedRoute>} />
      <Route path="/reqs/new" element={<ProtectedRoute><ReqForm /></ProtectedRoute>} />
      <Route path="/reqs/:reqId" element={<ProtectedRoute><ReqDetail /></ProtectedRoute>} />
      <Route path="/reqs/:reqId/report" element={<ProtectedRoute><ReqReport /></ProtectedRoute>} />
      <Route path="/jobs" element={<ProtectedRoute><Jobs /></ProtectedRoute>} />
      <Route path="/jobs/new" element={<ProtectedRoute><JobForm /></ProtectedRoute>} />
      <Route path="/jobs/:jobId" element={<ProtectedRoute><JobDetail /></ProtectedRoute>} />
      <Route path="/jobs/:jobId/edit" element={<ProtectedRoute><JobForm /></ProtectedRoute>} />
      <Route path="/jobs/:jobId/pipeline" element={<ProtectedRoute><JobPipeline /></ProtectedRoute>} />
      <Route path="/jobs/:jobId/pipeline/rejected" element={<ProtectedRoute><RejectedList /></ProtectedRoute>} />
      <Route path="/candidates" element={<ProtectedRoute><Candidates /></ProtectedRoute>} />
      <Route path="/candidates/:candidateId" element={<ProtectedRoute><CandidateDetail /></ProtectedRoute>} />
      <Route path="/applications/:appId" element={<ProtectedRoute><ApplicationDetailPage /></ProtectedRoute>} />
      <Route path="/interviews" element={<ProtectedRoute><Interviews /></ProtectedRoute>} />
      <Route path="/referrals" element={<ProtectedRoute><Referrals /></ProtectedRoute>} />
      <Route path="/settings" element={<ProtectedRoute><Settings /></ProtectedRoute>} />
      <Route path="/team" element={<ProtectedRoute><Team /></ProtectedRoute>} />
      <Route path="/account" element={<ProtectedRoute><Account /></ProtectedRoute>} />

      {/* Public routes (no auth) */}
      <Route path="/schedule/:token" element={<ScheduleBooking />} />
      <Route path="/careers/:orgSlug" element={<CareersPage />} />
      <Route path="/careers/:orgSlug/jobs/:jobId" element={<CareerJobDetail />} />
      <Route path="/careers/:orgSlug/jobs/:jobId/apply" element={<ApplyForm />} />

      <Route path="/" element={<DefaultRedirect />} />
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
