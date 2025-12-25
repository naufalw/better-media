import * as React from "react";
import { Link, Outlet, createRootRoute } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import {
  Film,
  Radio,
  Upload,
  Tv,
  Settings,
  LayoutDashboard,
  ChevronLeft,
  Menu,
} from "lucide-react";
import { useState } from "react";

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
  const [isCollapsed, setIsCollapsed] = useState(false);

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
            {!isCollapsed && <span>Settings</span>}
          </Link>
        </div>
      </aside>

      {/* Main Content */}
      <main
        className={`flex-1 min-w-0 transition-all duration-300 ease-in-out bg-[#0c0c0e] ${
          isCollapsed ? "ml-16" : "ml-64"
        }`}
      >
        <div className="max-w-7xl mx-auto p-8">
          <Outlet />
        </div>
      </main>

      <TanStackRouterDevtools position="bottom-right" />
    </div>
  );
}
