import * as React from "react";
import { Link, Outlet, createRootRoute, useLocation, useNavigate } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import {
  Film,
  Radio,
  Tv,
  Settings,
  LayoutDashboard,
  ChevronLeft,
  Menu,
  Loader2,
  Users,
} from "lucide-react";
import { useState, useEffect } from "react";
import { AuthProvider, useAuth } from "../lib/auth";

export const Route = createRootRoute({
  component: RootComponent,
});

const navItems = [
  { to: "/", icon: LayoutDashboard, label: "Dashboard" },
  { to: "/video", icon: Film, label: "Videos" },
  { to: "/streams", icon: Radio, label: "Stream Keys" },
  { to: "/live", icon: Tv, label: "Live Now" },
];

function RootComponent() {
  return (
    <AuthProvider>
      <AppContent />
      <TanStackRouterDevtools position="bottom-right" />
    </AuthProvider>
  );
}

function AppContent() {
  const [isCollapsed, setIsCollapsed] = useState(false);
  const { user, isLoading } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();

  const isPublicRoute = ["/login", "/setup"].includes(location.pathname);

  useEffect(() => {
    if (!isLoading) {
      if (!user && !isPublicRoute) {
        navigate({ to: "/login" });
      } else if (user && isPublicRoute) {
        navigate({ to: "/" });
      }
    }
  }, [user, isLoading, location.pathname, navigate, isPublicRoute]);

  if (isLoading) {
    return (
      <div className="min-h-screen bg-[#09090b] flex items-center justify-center text-zinc-200">
        <Loader2 className="w-6 h-6 animate-spin text-emerald-500" />
      </div>
    );
  }

  if (isPublicRoute) {
    return <Outlet />;
  }

  if (!user) return null; // Should redirect

  return (
    <div className="min-h-screen bg-[#09090b] text-zinc-200 flex font-sans antialiased selection:bg-emerald-500/30">
      {/* Sidebar */}
      <aside
        className={`bg-[#050505] border-r border-[#1f1f1f] flex flex-col fixed inset-y-0 left-0 z-50 transition-all duration-300 ease-in-out ${
          isCollapsed ? "w-16" : "w-64"
        }`}
      >
        {/* Header */}
        <div className="h-14 flex items-center justify-between px-4 border-b border-[#1f1f1f]">
          {!isCollapsed && (
            <div className="flex items-center gap-2 font-semibold text-white tracking-tight">
              <div className="p-1 bg-emerald-500 text-black">
                <LayoutDashboard className="w-4 h-4" />
              </div>
              <span className="text-sm">Better Media</span>
            </div>
          )}
          <button
            onClick={() => setIsCollapsed(!isCollapsed)}
            className={`p-1.5 hover:bg-zinc-900 text-zinc-500 hover:text-white transition-colors ${
              isCollapsed ? "mx-auto" : ""
            }`}
          >
            {isCollapsed ? <Menu className="w-4 h-4" /> : <ChevronLeft className="w-4 h-4" />}
          </button>
        </div>

        {/* Navigation */}
        <nav className="flex-1 py-6 space-y-1">
          {!isCollapsed && (
            <div className="mb-2 px-6 text-[10px] font-bold text-zinc-600 uppercase tracking-widest">
              Platform
            </div>
          )}
          {navItems.map((item) => (
            <Link
              key={item.to}
              to={item.to}
              className={`flex items-center gap-3 px-4 py-2 text-sm font-medium text-zinc-400 hover:text-white hover:bg-zinc-900/50 transition-colors border-l-2 border-transparent group ${
                isCollapsed ? "justify-center px-0 py-3 mx-2" : "mx-0"
              }`}
              activeProps={{
                className: `!text-emerald-500 !bg-emerald-500/5 ${
                  isCollapsed ? "!border-l-0" : "!border-l-emerald-500"
                }`,
              }}
              activeOptions={{ exact: item.to === "/" }}
              title={isCollapsed ? item.label : undefined}
            >
              <item.icon
                className={`w-4 h-4 group-hover:text-white transition-colors ${
                  isCollapsed ? "" : ""
                }`}
              />
              {!isCollapsed && <span>{item.label}</span>}
            </Link>
          ))}

          {/* Admin Section */}
          {user.role === "admin" && (
            <>
              <div
                className={`mt-6 mb-2 px-6 text-[10px] font-bold text-zinc-600 uppercase tracking-widest ${isCollapsed ? "hidden" : ""}`}
              >
                Admin
              </div>
              <Link
                to="/users"
                className={`flex items-center gap-3 px-4 py-2 text-sm font-medium text-zinc-400 hover:text-white hover:bg-zinc-900/50 transition-colors border-l-2 border-transparent group ${
                  isCollapsed ? "justify-center px-0 py-3 mx-2" : "mx-0"
                }`}
                activeProps={{
                  className: `!text-emerald-500 !bg-emerald-500/5 ${
                    isCollapsed ? "!border-l-0" : "!border-l-emerald-500"
                  }`,
                }}
                title={isCollapsed ? "Users" : undefined}
              >
                <Users className="w-4 h-4 group-hover:text-white transition-colors" />
                {!isCollapsed && <span>Users</span>}
              </Link>
            </>
          )}
        </nav>

        {/* Footer Navigation */}
        <div className="p-2 border-t border-[#1f1f1f]">
          <Link
            to="/settings"
            className={`flex items-center gap-3 px-4 py-2 text-sm font-medium text-zinc-400 hover:text-white hover:bg-zinc-900/50 transition-colors ${
              isCollapsed ? "justify-center px-0" : ""
            }`}
            activeProps={{
              className: "!text-white !bg-zinc-900",
            }}
            title="Settings"
          >
            <Settings className="w-4 h-4" />
            {!isCollapsed && (
              <div className="flex flex-col items-start overflow-hidden">
                <span className="truncate w-full">{user.name}</span>
                <span className="text-xs text-zinc-500">Settings</span>
              </div>
            )}
          </Link>
        </div>
      </aside>

      {/* Main Content */}
      <main
        className={`flex-1 min-w-0 transition-all duration-300 ease-in-out bg-[#0c0c0e] ${
          isCollapsed ? "ml-16" : "ml-64"
        }`}
      >
        <div className="h-full w-full">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
