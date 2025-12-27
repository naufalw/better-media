import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_padded/")({
  component: DashboardPage,
});

function DashboardPage() {
  return (
    <div className="">
      <h1 className="text-2xl font-bold text-white">Dashboard</h1>
      <p className="text-zinc-500">Welcome to Better Media.</p>
    </div>
  );
}
