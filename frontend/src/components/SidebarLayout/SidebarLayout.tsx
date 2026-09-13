import { cn } from "@heroui/react";
import type { ReactNode } from "react";

/**
 * Sidebar layouts properties.
 */
type Props = {
  /** The sidebar component, will be rendered inside a `aside` tag */
  sidebar?: ReactNode;

  /** Children are the main content in the layout */
  children?: ReactNode;

  /** Controls if the sidebar is open */
  isOpen?: boolean;
};

/**
 * Renders content together a sidebar.
 */
function SidebarLayout({ children, sidebar, isOpen }: Props) {
  return (
    <div className="w-full h-full flex flex-row relative">
      <span className={cn("transition-[width]", isOpen ? "w-1/5" : "w-0")} />
      <aside className="min-w-0 w-1/5 overflow-hidden  transition-[width] absolute top-0 left-0 h-full">
        <div
          className={cn(
            "h-full w-full transition-transform border-r border-r-neutral-200",
            isOpen ? "translate-x-0" : "-translate-x-full",
          )}
        >
          {sidebar}
        </div>
      </aside>
      <div className="h-full flex-1">{children}</div>
    </div>
  );
}

export default SidebarLayout;
