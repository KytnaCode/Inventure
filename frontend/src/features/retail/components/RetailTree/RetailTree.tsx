import { useMemo, useState } from "react";
import type { Place, Retail } from "../../model/model";
import type { TreeData } from "../../../../components/TreeView/TreeView";
import { Typography } from "@heroui/react";
import TreeView from "../../../../components/TreeView/TreeView";

type Props = {
  retails?: Retail[];
};

function RetailTree({ retails }: Props) {
  const [isOpen, setIsOpen] = useState(new Map());

  const handleClick = (path: TreeData[]) => {
    setIsOpen((m) =>
      new Map(m).set(
        path[path.length - 1].id,
        !m.get(path[path.length - 1].id),
      ),
    );
  };

  const data = useMemo(() => {
    function convertPlaces(p: Place): TreeData {
      const children = p.Children?.map(convertPlaces);

      return {
        id: p.ID,
        value: <Typography type="body">{p.Name}</Typography>,
        isOpen: !!isOpen.get(p.ID),
        children,
      };
    }

    const root: TreeData = {
      id: crypto.randomUUID(),
      value: "root",
      isOpen: true,
      children: retails?.map((r) => ({
        id: r.ID,
        value: <Typography type="body">{r.Name}</Typography>,
        isOpen: !!isOpen.get(r.ID),
        children: r.Storage.Children?.map(convertPlaces),
      })),
    };

    return root;
  }, [retails, isOpen]);

  return <TreeView hideRoot onClick={handleClick} tree={data} />;
}

export default RetailTree;
