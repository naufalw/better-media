import { PaddedLayout } from "@/components/padded-layout";
import { createFileRoute, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute("/_padded")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <PaddedLayout>
      <Outlet />
    </PaddedLayout>
  );
}
