import { ReactNode } from "react";

interface PaddedLayoutProps {
  children: ReactNode;
}

export function PaddedLayout({ children }: PaddedLayoutProps) {
  return <div className="max-w-7xl mx-auto p-8 h-full">{children}</div>;
}
