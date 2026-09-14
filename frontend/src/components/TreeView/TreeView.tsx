import { Surface } from "@heroui/react";
import { createContext, use, type ReactNode } from "react";
import { cn } from "../../util/cn";

type Props = {
  tree?: TreeData;
  selectedID?: string;
  onClick?: (path: TreeData[]) => void;
  hideRoot?: boolean;
};

type NodeProps = {
  tree?: TreeData;
  path: TreeData[];
  hide?: boolean;
};

export type TreeData = {
  id: string;
  value: ReactNode;
  isOpen?: boolean;
  children?: TreeData[];
};

type TreeContextData = {
  selectedID?: string;
  onClick?: (path: TreeData[]) => void;
};

const TreeContext = createContext<TreeContextData>({});

function TreeView({ tree, selectedID, onClick, hideRoot }: Props) {
  const data: TreeContextData = {
    selectedID,
    onClick,
  };

  return (
    <TreeContext.Provider value={data}>
      <TreeNode tree={tree} path={[]} hide={hideRoot} />
    </TreeContext.Provider>
  );
}

function TreeNode({ tree, path, hide }: NodeProps) {
  const { selectedID, onClick } = use(TreeContext);

  const handleClick = () => {
    if (tree && onClick) {
      onClick([...path, tree]);
    }
  };

  return (
    tree && (
      <div>
        <Surface
          onClick={handleClick}
          className={cn(
            hide && "hidden",
            "hover:bg-surface-secondary/40",
            "w-full p-1 rounded-lg",
            selectedID && tree?.id === selectedID
              ? "bg-surface-secondary/70"
              : "bg-surface",
          )}
        >
          {tree?.value}
        </Surface>

        {tree?.children && (
          <div className="h-fit w-full relative">
            <div
              className={cn(
                "flex flex-row w-full overflow-hidden",
                tree.isOpen === false ? "h-0" : "h-full",
              )}
            >
              <span className="w-4 h-0" />
              <div className="w-0 h-full border-r absolute top-0 left-1 bg-accent" />
              <div
                className={cn(
                  "transition-transform",
                  tree.isOpen === false ? "-translate-y-full" : "translate-y-0",
                )}
              >
                <div>
                  <div className="flex flex-col w-full">
                    {tree.children.map((child) => (
                      <TreeNode
                        tree={child}
                        key={child.id}
                        path={[...path, tree]}
                      />
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    )
  );
}

export default TreeView;
