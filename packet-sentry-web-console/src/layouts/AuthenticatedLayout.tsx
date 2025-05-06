import { Outlet, useLocation, useNavigate } from "react-router-dom";
import { AppSidebar } from "@/components/AppSidebar";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Separator } from "@/components/ui/separator";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Cog, UserCircle } from "lucide-react";
import { ModeToggle } from "@/components/ModeToggle";
import { Button } from "@/components/ui/button";
import { generateBreadcrumbs } from "@/contexts/BreadcrumbContext";
import React from "react";

export default function AuthenticatedLayout() {
  const location = useLocation();
  const navigate = useNavigate();

  const navData = {
    navMain: [
      {
        title: "Device Management",
        items: [
          { title: "New Device", url: "/devices/new" },
          { title: "Existing Devices", url: "/devices/list" },
        ],
      },
    ],
  };

  const breadcrumbs = generateBreadcrumbs(location.pathname);

  const handleMenuClick = (item: string) => {
    navigate(item);
  };

  return (
    <SidebarProvider>
      <AppSidebar navData={navData} />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 h-4" />
          <div className="w-full px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
            <Breadcrumb className="flex items-center space-x-2">
              <BreadcrumbList>
                {breadcrumbs.map((crumb, idx) => (
                  <React.Fragment key={idx}>
                    <BreadcrumbItem key={idx}>
                      {crumb.href ? (
                        <BreadcrumbLink href={crumb.href}>
                          {crumb.label}
                        </BreadcrumbLink>
                      ) : (
                        <BreadcrumbPage>{crumb.label}</BreadcrumbPage>
                      )}
                    </BreadcrumbItem>
                    {idx < breadcrumbs.length - 1 && <BreadcrumbSeparator />}
                  </React.Fragment>
                ))}
              </BreadcrumbList>
            </Breadcrumb>
            <div className="flex items-center space-x-4">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" size="icon">
                    <Cog className="w-5 h-5" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem onClick={() => handleMenuClick("/billing")}>
                    Billing Details
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => handleMenuClick("/administrators")}
                  >
                    Administrators
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" size="icon">
                    <UserCircle className="w-5 h-5" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem
                    onClick={() => handleMenuClick("/settings")}
                  >
                    Settings
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => handleMenuClick("/reset-password")}
                  >
                    Reset Password
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => handleMenuClick("/logout")}>
                    Logout
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>

              <ModeToggle />
            </div>
          </div>
        </header>
        <main className="flex flex-1 flex-col gap-4 p-4">
          <Outlet />
        </main>
      </SidebarInset>
    </SidebarProvider>
  );
}
