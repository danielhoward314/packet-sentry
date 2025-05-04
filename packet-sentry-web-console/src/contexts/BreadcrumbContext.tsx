const breadcrumbMap: Record<string, { label: string; href?: string }[]> = {
  '/administrators': [{ label: 'Administrators', href: '/administrators' }],
  '/billing': [{ label: 'Billing Details', href: '/billing' }],
  '/devices/list': [
    { label: 'Device Management', href: '/devices/list' },
    { label: 'Existing Devices' },
  ],
  '/devices/new': [
    { label: 'Device Management', href: '/devices/new' },
    { label: 'New Device' },
  ],
  '/logout': [{ label: 'Logout', href: '/logout' }],
  '/settings': [{ label: 'Settings', href: '/settings' }],
}

export function generateBreadcrumbs(pathname: string) {
  return breadcrumbMap[pathname] ?? [{ label: 'Home', href: '/' }]
}
