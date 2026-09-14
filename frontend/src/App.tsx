import { useState } from "react";
import TreeView, { type TreeData } from "./components/TreeView/TreeView";
import { produce } from "immer";

const tree: TreeData = {
  id: "a",
  value: "a",
  children: [
    {
      id: "aa",
      value: "aa",
    },
    {
      id: "ab",
      value: "ab",
      isOpen: false,
      children: [
        {
          id: "aba",
          value: "aba",
        },
        {
          id: "abb",
          value: "abb",
        },
      ],
    },
    {
      id: "ac",
      value: "ac",
    },
  ],
};

function App() {
  const [treeData, setTreeData] = useState(tree);

  return (
    <div className="w-full h-full place-items-center grid">
      <TreeView
        tree={treeData}
        selectedID="ab"
        onClick={(path) => {
          setTreeData((d) =>
            produce(d, (draft) => {
              for (let i = 1; i < path.length; i++) {
                const index = draft.children?.findIndex(
                  (e) => e.id === path[i].id,
                );

                if (index === undefined) {
                  return;
                }

                if (draft.children === undefined) {
                  return;
                }

                draft = draft.children[index];
              }

              if (draft === undefined) {
                throw "errro";
              }

              draft.isOpen = !draft.isOpen;
            }),
          );
        }}
      />
    </div>
  );
}

export default App;
