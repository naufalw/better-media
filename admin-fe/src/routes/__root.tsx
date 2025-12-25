import * as React from "react";
import { Link, Outlet, createRootRoute } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { Film, Radio, Upload, Tv, Settings, LayoutDashboard } from "lucide-react";

export const Route = createRootRoute({
  component: RootComponent,
});

const navItems = [
  { to: "/", icon: Film, label: "Videos" },
  { to: "/upload", icon: Upload, label: "Upload" },
  { to: "/streams", icon: Radio, label: "Stream Keys" },
  { to: "/live", icon: Tv, label: "Live Now" },
];

function RootComponent() {
  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-200 flex font-sans antialiased selection:bg-zinc-800">
      {/* Sidebar */}
      <aside className="w-64 bg-zinc-950 border-r border-zinc-900 flex flex-col fixed inset-y-0 left-0 z-50">
        {/* Header */}
        <div className="h-16 flex items-center px-6 border-b border-zinc-900">
          <div className="flex items-center gap-2 font-semibold text-white tracking-tight">
            <LayoutDashboard className="w-5 h-5 text-zinc-400" />
            <span>Better Media</span>
          </div>
        </div>

        {/* Navigation */}
        <nav className="flex-1 p-4 space-y-1">
          <div className="mb-2 px-2 text-xs font-medium text-zinc-500 uppercase tracking-wider">
            Platform
          </div>
          {navItems.map((item) => (
            <Link
              key={item.to}
              to={item.to}
              className="flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium text-zinc-400 hover:text-white hover:bg-zinc-900/50 transition-colors"
              activeProps={{
                className: "!text-white !bg-zinc-900",
              }}
              activeOptions={{ exact: item.to === "/" }}
            >
              <item.icon className="w-4 h-4" />
              <span>{item.label}</span>
            </Link>
          ))}
        </nav>

        {/* Footer Navigation */}
        <div className="p-4 border-t border-zinc-900">
          <Link
            to="/settings"
            className="flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium text-zinc-400 hover:text-white hover:bg-zinc-900/50 transition-colors"
            activeProps={{
              className: "!text-white !bg-zinc-900",
            }}
          >
            <Settings className="w-4 h-4" />
            <span>Settings</span>
          </Link>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 ml-64 min-w-0">
        <div className="max-w-6xl mx-auto p-8">
          <Outlet />
        </div>
      </main>

      <TanStackRouterDevtools position="bottom-right" />
    </div>
  );
}
