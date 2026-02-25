import { useState, useEffect } from 'react'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'
import { Card } from '@/components/common/Card'
import { Button } from '@/components/common/Button'

interface ScheduledReport {
  id: number
  shop_id: number
  report_type: 'daily' | 'weekly' | 'monthly'
  schedule_time: string
  schedule_day?: number
  recipients: string[]
  is_active: boolean
  last_sent_at?: string
  created_at: string
}

export default function ScheduledReports() {
  const shop = useAuthStore((state) => state.shop)
  const [reports, setReports] = useState<ScheduledReport[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [editingReport, setEditingReport] = useState<ScheduledReport | null>(null)
  const [form, setForm] = useState({
    report_type: 'daily' as 'daily' | 'weekly' | 'monthly',
    schedule_time: '08:00',
    schedule_day: 1,
    recipients: '',
    is_active: true
  })

  useEffect(() => {
    if (shop?.id) {
      fetchScheduledReports()
    }
  }, [shop?.id])

  const fetchScheduledReports = async () => {
    if (!shop?.id) return
    try {
      const response = await api.get('/v1/reports/scheduled')
      const data = response.data?.data || response.data || []
      setReports(Array.isArray(data) ? data : [])
    } catch (e) {
      console.error('Failed to fetch scheduled reports:', e)
    } finally {
      setIsLoading(false)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!shop?.id) return

    const payload = {
      shop_id: shop.id,
      report_type: form.report_type,
      schedule_time: form.schedule_time,
      schedule_day: form.report_type === 'monthly' ? form.schedule_day : undefined,
      recipients: form.recipients.split(',').map(r => r.trim()).filter(Boolean),
      is_active: form.is_active
    }

    try {
      if (editingReport) {
        await api.put(`/v1/reports/scheduled/${editingReport.id}`, payload)
      } else {
        await api.post('/v1/reports/scheduled', payload)
      }
      setShowModal(false)
      setEditingReport(null)
      setForm({
        report_type: 'daily',
        schedule_time: '08:00',
        schedule_day: 1,
        recipients: '',
        is_active: true
      })
      fetchScheduledReports()
    } catch (e) {
      console.error('Failed to save scheduled report:', e)
    }
  }

  const handleToggle = async (report: ScheduledReport) => {
    try {
      await api.put(`/v1/reports/scheduled/${report.id}`, {
        is_active: !report.is_active
      })
      fetchScheduledReports()
    } catch (e) {
      console.error('Failed to toggle report:', e)
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this scheduled report?')) return
    try {
      await api.delete(`/v1/reports/scheduled/${id}`)
      fetchScheduledReports()
    } catch (e) {
      console.error('Failed to delete report:', e)
    }
  }

  const openEditModal = (report: ScheduledReport) => {
    setEditingReport(report)
    setForm({
      report_type: report.report_type,
      schedule_time: report.schedule_time,
      schedule_day: report.schedule_day || 1,
      recipients: report.recipients.join(', '),
      is_active: report.is_active
    })
    setShowModal(true)
  }

  const getReportTypeLabel = (type: string) => {
    const labels: Record<string, string> = {
      daily: 'Daily Report',
      weekly: 'Weekly Report',
      monthly: 'Monthly Report'
    }
    return labels[type] || type
  }

  const getDayLabel = (day: number) => {
    const days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']
    return days[day] || ''
  }

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="h-8 w-48 bg-gray-200 animate-pulse rounded"></div>
        <div className="space-y-3">
          <div className="h-24 bg-gray-200 animate-pulse rounded-xl"></div>
          <div className="h-24 bg-gray-200 animate-pulse rounded-xl"></div>
        </div>
      </div>
    )
  }

  return (
    <div>
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Scheduled Reports</h1>
          <p className="text-gray-500 mt-1">Automatically receive reports via WhatsApp or Email</p>
        </div>
        <Button onClick={() => setShowModal(true)} leftIcon={
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
        }>
          Add Schedule
        </Button>
      </div>

      {reports.length === 0 ? (
        <Card>
          <div className="text-center py-12">
            <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <svg className="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">No Scheduled Reports</h3>
            <p className="text-gray-500 mb-4">Set up automatic reports to receive updates daily, weekly, or monthly</p>
            <Button onClick={() => setShowModal(true)}>Create First Schedule</Button>
          </div>
        </Card>
      ) : (
        <div className="space-y-4">
          {reports.map((report) => (
            <Card key={report.id} className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${
                  report.is_active ? 'bg-primary-50' : 'bg-gray-100'
                }`}>
                  <svg className={`w-6 h-6 ${report.is_active ? 'text-primary' : 'text-gray-400'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                </div>
                <div>
                  <h3 className="font-semibold text-gray-900">{getReportTypeLabel(report.report_type)}</h3>
                  <p className="text-sm text-gray-500">
                    {report.report_type === 'daily' && `Every day at ${report.schedule_time}`}
                    {report.report_type === 'weekly' && `Every ${getDayLabel(new Date(report.schedule_time).getDay())} at ${report.schedule_time}`}
                    {report.report_type === 'monthly' && `Day ${report.schedule_day} of each month at ${report.schedule_time}`}
                  </p>
                  <p className="text-sm text-gray-500">Recipients: {report.recipients.join(', ')}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                {report.last_sent_at && (
                  <span className="text-xs text-gray-500">
                    Last sent: {new Date(report.last_sent_at).toLocaleDateString()}
                  </span>
                )}
                <button
                  onClick={() => handleToggle(report)}
                  className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                    report.is_active ? 'bg-primary' : 'bg-gray-200'
                  }`}
                >
                  <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                    report.is_active ? 'translate-x-6' : 'translate-x-1'
                  }`} />
                </button>
                <button
                  onClick={() => openEditModal(report)}
                  className="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                  </svg>
                </button>
                <button
                  onClick={() => handleDelete(report.id)}
                  className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </Card>
          ))}
        </div>
      )}

      {/* Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <Card className="w-full max-w-md">
            <div className="p-4 border-b border-gray-100">
              <h3 className="text-xl font-bold text-gray-900">
                {editingReport ? 'Edit Scheduled Report' : 'Schedule New Report'}
              </h3>
            </div>
            <form onSubmit={handleSubmit} className="p-4 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Report Type</label>
                <select
                  value={form.report_type}
                  onChange={(e) => setForm({ ...form, report_type: e.target.value as any })}
                  className="w-full px-4 py-3 border rounded-xl"
                >
                  <option value="daily">Daily Report</option>
                  <option value="weekly">Weekly Report</option>
                  <option value="monthly">Monthly Report</option>
                </select>
              </div>

              {form.report_type === 'monthly' && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Day of Month</label>
                  <select
                    value={form.schedule_day}
                    onChange={(e) => setForm({ ...form, schedule_day: parseInt(e.target.value) })}
                    className="w-full px-4 py-3 border rounded-xl"
                  >
                    {Array.from({ length: 28 }, (_, i) => i + 1).map(day => (
                      <option key={day} value={day}>{day}</option>
                    ))}
                  </select>
                </div>
              )}

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Time</label>
                <input
                  type="time"
                  value={form.schedule_time}
                  onChange={(e) => setForm({ ...form, schedule_time: e.target.value })}
                  className="w-full px-4 py-3 border rounded-xl"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Recipients (comma separated)</label>
                <input
                  type="text"
                  value={form.recipients}
                  onChange={(e) => setForm({ ...form, recipients: e.target.value })}
                  placeholder="phone1, email1, phone2"
                  className="w-full px-4 py-3 border rounded-xl"
                />
                <p className="text-xs text-gray-500 mt-1">Phone numbers or email addresses</p>
              </div>

              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="checkbox"
                  checked={form.is_active}
                  onChange={(e) => setForm({ ...form, is_active: e.target.checked })}
                  className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
                />
                <span className="text-sm text-gray-700">Active</span>
              </label>

              <div className="flex gap-3 pt-4">
                <Button type="button" variant="secondary" onClick={() => {
                  setShowModal(false)
                  setEditingReport(null)
                }} className="flex-1">
                  Cancel
                </Button>
                <Button type="submit" className="flex-1">
                  {editingReport ? 'Update' : 'Create'}
                </Button>
              </div>
            </form>
          </Card>
        </div>
      )}
    </div>
  )
}
