import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/")({
  component: DashboardPage,
});

function DashboardPage() {
  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold text-white">Dashboard</h1>
      <p className="text-zinc-500">Welcome to Better Media.</p>
    </div>
  );
}
