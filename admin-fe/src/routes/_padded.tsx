import { createFileRoute, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute("/_padded")({
  component: LayoutComponent,
});

function LayoutComponent() {
  return (
    <div className="max-w-7xl mx-auto p-8 h-full">
      <Outlet />
    </div>
  );
}
