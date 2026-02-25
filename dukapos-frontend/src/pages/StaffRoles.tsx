import { useState, useEffect } from 'react'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'
import { Card } from '@/components/common/Card'
import { Button } from '@/components/common/Button'
import { Badge } from '@/components/common/Badge'

export interface Role {
  id: number
  name: string
  description: string
  permissions: string[]
  is_default: boolean
  staff_count: number
}

export interface Permission {
  id: string
  name: string
  category: string
  description: string
}

const PERMISSIONS: Permission[] = [
  { id: 'products.view', name: 'View Products', category: 'Products', description: 'Can view product list' },
  { id: 'products.create', name: 'Create Products', category: 'Products', description: 'Can add new products' },
  { id: 'products.edit', name: 'Edit Products', category: 'Products', description: 'Can edit product details' },
  { id: 'products.delete', name: 'Delete Products', category: 'Products', description: 'Can delete products' },
  { id: 'sales.view', name: 'View Sales', category: 'Sales', description: 'Can view sales records' },
  { id: 'sales.create', name: 'Create Sales', category: 'Sales', description: 'Can process new sales' },
  { id: 'sales.refund', name: 'Refund Sales', category: 'Sales', description: 'Can process refunds' },
  { id: 'customers.view', name: 'View Customers', category: 'Customers', description: 'Can view customer list' },
  { id: 'customers.create', name: 'Add Customers', category: 'Customers', description: 'Can add new customers' },
  { id: 'customers.edit', name: 'Edit Customers', category: 'Customers', description: 'Can edit customer details' },
  { id: 'reports.view', name: 'View Reports', category: 'Reports', description: 'Can view reports' },
  { id: 'reports.export', name: 'Export Reports', category: 'Reports', description: 'Can export report data' },
  { id: 'settings.view', name: 'View Settings', category: 'Settings', description: 'Can view shop settings' },
  { id: 'settings.edit', name: 'Edit Settings', category: 'Settings', description: 'Can modify shop settings' },
  { id: 'staff.manage', name: 'Manage Staff', category: 'Staff', description: 'Can add/remove staff' },
  { id: 'billing.view', name: 'View Billing', category: 'Billing', description: 'Can view billing info' },
  { id: 'billing.manage', name: 'Manage Billing', category: 'Billing', description: 'Can manage subscription' },
]

const DEFAULT_ROLES: Omit<Role, 'id' | 'staff_count'>[] = [
  {
    name: 'Admin',
    description: 'Full access to all features',
    permissions: PERMISSIONS.map(p => p.id),
    is_default: false
  },
  {
    name: 'Manager',
    description: 'Can manage products, sales, and reports',
    permissions: [
      'products.view', 'products.create', 'products.edit',
      'sales.view', 'sales.create',
      'customers.view', 'customers.create', 'customers.edit',
      'reports.view', 'reports.export',
      'settings.view'
    ],
    is_default: false
  },
  {
    name: 'Cashier',
    description: 'Can process sales and view products',
    permissions: [
      'products.view',
      'sales.view', 'sales.create',
      'customers.view', 'customers.create'
    ],
    is_default: false
  },
  {
    name: 'Viewer',
    description: 'Read-only access',
    permissions: [
      'products.view',
      'sales.view',
      'customers.view',
      'reports.view'
    ],
    is_default: true
  }
]

export default function StaffRoles() {
  const shop = useAuthStore((state) => state.shop)
  const [roles, setRoles] = useState<Role[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [editingRole, setEditingRole] = useState<Role | null>(null)
  const [form, setForm] = useState({
    name: '',
    description: '',
    permissions: [] as string[],
    is_default: false
  })
  const [activeTab, setActiveTab] = useState<'roles' | 'permissions'>('roles')

  useEffect(() => {
    if (shop?.id) {
      fetchRoles()
    }
  }, [shop?.id])

  const fetchRoles = async () => {
    if (!shop?.id) return
    try {
      const response = await api.get('/v1/staff/roles')
      const data = response.data?.data || response.data || []
      if (Array.isArray(data) && data.length > 0) {
        setRoles(data)
      } else {
        setRoles(DEFAULT_ROLES.map((r, i) => ({
          ...r,
          id: i + 1,
          staff_count: 0
        })))
      }
    } catch (e) {
      setRoles(DEFAULT_ROLES.map((r, i) => ({
        ...r,
        id: i + 1,
        staff_count: 0
      })))
    } finally {
      setIsLoading(false)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!shop?.id) return

    try {
      if (editingRole) {
        await api.put(`/v1/staff/roles/${editingRole.id}`, {
          ...form,
          shop_id: shop.id
        })
      } else {
        await api.post('/v1/staff/roles', {
          ...form,
          shop_id: shop.id
        })
      }
      setShowModal(false)
      resetForm()
      fetchRoles()
    } catch (e) {
      console.error('Failed to save role:', e)
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this role?')) return
    try {
      await api.delete(`/v1/staff/roles/${id}`)
      fetchRoles()
    } catch (e) {
      console.error('Failed to delete role:', e)
    }
  }

  const resetForm = () => {
    setEditingRole(null)
    setForm({
      name: '',
      description: '',
      permissions: [],
      is_default: false
    })
  }

  const openEditModal = (role: Role) => {
    setEditingRole(role)
    setForm({
      name: role.name,
      description: role.description,
      permissions: role.permissions,
      is_default: role.is_default
    })
    setShowModal(true)
  }

  const togglePermission = (permId: string) => {
    setForm(prev => ({
      ...prev,
      permissions: prev.permissions.includes(permId)
        ? prev.permissions.filter(p => p !== permId)
        : [...prev.permissions, permId]
    }))
  }

  const getPermissionsByCategory = () => {
    const categories: Record<string, Permission[]> = {}
    PERMISSIONS.forEach(p => {
      if (!categories[p.category]) categories[p.category] = []
      categories[p.category].push(p)
    })
    return categories
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
          <h1 className="text-2xl font-bold text-gray-900">Staff Roles & Permissions</h1>
          <p className="text-gray-500 mt-1">Manage staff access levels</p>
        </div>
        <Button onClick={() => { resetForm(); setShowModal(true); }}>
          Create Role
        </Button>
      </div>

      {/* Tabs */}
      <div className="flex gap-2 mb-6 p-1 bg-gray-100 rounded-xl w-fit">
        <button
          onClick={() => setActiveTab('roles')}
          className={`py-2 px-4 rounded-lg font-medium text-sm transition ${
            activeTab === 'roles' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-500 hover:text-gray-700'
          }`}
        >
          Roles ({roles.length})
        </button>
        <button
          onClick={() => setActiveTab('permissions')}
          className={`py-2 px-4 rounded-lg font-medium text-sm transition ${
            activeTab === 'permissions' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-500 hover:text-gray-700'
          }`}
        >
          All Permissions
        </button>
      </div>

      {activeTab === 'roles' ? (
        <div className="space-y-4">
          {roles.map((role) => (
            <Card key={role.id} className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <div className="w-12 h-12 bg-primary-50 rounded-xl flex items-center justify-center">
                  <svg className="w-6 h-6 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
                  </svg>
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="font-semibold text-gray-900">{role.name}</h3>
                    {role.is_default && (
                      <Badge variant="info">Default</Badge>
                    )}
                  </div>
                  <p className="text-sm text-gray-500">{role.description}</p>
                  <p className="text-xs text-gray-400">{role.staff_count} staff members</p>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <div className="flex flex-wrap gap-1 max-w-xs">
                  {role.permissions.slice(0, 3).map(p => (
                    <span key={p} className="text-xs bg-gray-100 text-gray-600 px-2 py-1 rounded">
                      {p.split('.')[1]}
                    </span>
                  ))}
                  {role.permissions.length > 3 && (
                    <span className="text-xs text-gray-400">+{role.permissions.length - 3} more</span>
                  )}
                </div>
                <button
                  onClick={() => openEditModal(role)}
                  className="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                  </svg>
                </button>
                {!role.is_default && (
                  <button
                    onClick={() => handleDelete(role.id)}
                    className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg"
                  >
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                )}
              </div>
            </Card>
          ))}
        </div>
      ) : (
        <Card>
          <div className="divide-y divide-gray-100">
            {Object.entries(getPermissionsByCategory()).map(([category, perms]) => (
              <div key={category} className="py-4">
                <h3 className="font-semibold text-gray-900 mb-3">{category}</h3>
                <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-3">
                  {perms.map(perm => (
                    <label key={perm.id} className="flex items-start gap-3 cursor-pointer hover:bg-gray-50 p-2 rounded-lg">
                      <input
                        type="checkbox"
                        checked={form.permissions.includes(perm.id)}
                        onChange={() => togglePermission(perm.id)}
                        className="mt-1 w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
                      />
                      <div>
                        <p className="font-medium text-sm text-gray-900">{perm.name}</p>
                        <p className="text-xs text-gray-500">{perm.description}</p>
                      </div>
                    </label>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </Card>
      )}

      {/* Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <Card className="w-full max-w-lg max-h-[80vh] overflow-y-auto">
            <div className="p-4 border-b border-gray-100">
              <h3 className="text-xl font-bold text-gray-900">
                {editingRole ? 'Edit Role' : 'Create Role'}
              </h3>
            </div>
            <form onSubmit={handleSubmit} className="p-4 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Role Name</label>
                <input
                  type="text"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  className="w-full px-4 py-3 border rounded-xl"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Description</label>
                <textarea
                  value={form.description}
                  onChange={(e) => setForm({ ...form, description: e.target.value })}
                  className="w-full px-4 py-3 border rounded-xl"
                  rows={2}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Permissions</label>
                <div className="space-y-3 max-h-48 overflow-y-auto border rounded-xl p-3">
                  {Object.entries(getPermissionsByCategory()).map(([category, perms]) => (
                    <div key={category}>
                      <p className="text-xs font-semibold text-gray-500 uppercase mb-2">{category}</p>
                      <div className="space-y-2">
                        {perms.map((perm: { id: string; name: string }) => (
                          <label key={perm.id} className="flex items-center gap-2 cursor-pointer">
                            <input
                              type="checkbox"
                              checked={form.permissions.includes(perm.id)}
                              onChange={() => togglePermission(perm.id)}
                              className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
                            />
                            <span className="text-sm text-gray-700">{perm.name}</span>
                          </label>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
              <label className="flex items-center gap-3 cursor-pointer">
                <input
                  type="checkbox"
                  checked={form.is_default}
                  onChange={(e) => setForm({ ...form, is_default: e.target.checked })}
                  className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
                />
                <span className="text-sm text-gray-700">Set as default role for new staff</span>
              </label>
              <div className="flex gap-3 pt-4">
                <Button type="button" variant="secondary" onClick={() => { setShowModal(false); resetForm(); }} className="flex-1">
                  Cancel
                </Button>
                <Button type="submit" className="flex-1">
                  {editingRole ? 'Update' : 'Create'}
                </Button>
              </div>
            </form>
          </Card>
        </div>
      )}
    </div>
  )
}
