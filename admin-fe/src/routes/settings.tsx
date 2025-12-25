import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/settings")({
  component: SettingsPage,
});

function SettingsPage() {
  return (
    <div className="max-w-2xl space-y-8">
      <div>
        <h1 className="text-xl font-semibold text-white">System Settings</h1>
        <p className="text-sm text-zinc-500 mt-1">Configure global server parameters</p>
      </div>

      <div className="space-y-6">
        <div className="bg-zinc-900/30 border border-zinc-800 rounded-lg overflow-hidden">
          <div className="px-6 py-4 border-b border-zinc-800">
            <h3 className="text-sm font-medium text-white">Instance Information</h3>
          </div>
          <div className="p-6 space-y-4">
            <div className="grid grid-cols-3 gap-4 text-sm">
              <div className="text-zinc-500">API Version</div>
              <div className="col-span-2 text-zinc-300 font-mono">v1.0.0-beta</div>
            </div>
            <div className="grid grid-cols-3 gap-4 text-sm">
              <div className="text-zinc-500">Environment</div>
              <div className="col-span-2 text-zinc-300 font-mono">Production</div>
            </div>
            <div className="grid grid-cols-3 gap-4 text-sm">
              <div className="text-zinc-500">API Endpoint</div>
              <div className="col-span-2 text-zinc-300 font-mono">http://localhost:8080</div>
            </div>
            <div className="grid grid-cols-3 gap-4 text-sm">
              <div className="text-zinc-500">RTMP Server</div>
              <div className="col-span-2 text-zinc-300 font-mono">rtmp://localhost:1935/live</div>
            </div>
          </div>
        </div>

        <div className="bg-zinc-900/30 border border-zinc-800 rounded-lg overflow-hidden opacity-50">
          <div className="px-6 py-4 border-b border-zinc-800">
            <h3 className="text-sm font-medium text-white">Storage Configuration</h3>
          </div>
          <div className="p-6">
            <p className="text-sm text-zinc-500">
              S3 configuration is managed via environment variables.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
