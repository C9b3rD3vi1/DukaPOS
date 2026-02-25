import { Outlet } from 'react-router-dom'
import Header from './Header'
import Sidebar from './Sidebar'
import BottomNav from './BottomNav'
import { FAB } from '@/components/common/FAB'

export default function Layout() {
  return (
    <div className="min-h-screen bg-gray-50">
      <Header />
      <Sidebar />
      
      <main className="pt-16 md:pl-64 min-h-screen pb-20 md:pb-0">
        <div className="p-4 md:p-6 lg:p-8 max-w-7xl">
          <Outlet />
        </div>
      </main>
      
      <FAB to="/sales/new" />
      <BottomNav />
    </div>
  )
}
