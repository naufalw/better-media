import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/")({
  component: App,
});

function App() {
  return (
    <div className="bg-black h-screen">
      <h1>jfaijfiaj</h1>
    </div>
  );
}
