import { render, screen, waitFor } from "@testing-library/react";
import type { TreeData } from "./TreeView";
import TreeView from "./TreeView";
import userEvent from "@testing-library/user-event";
import { produce } from "immer";

const data: TreeData = {
  id: "r",
  value: "root",
  isOpen: true,
  children: [
    {
      id: "a",
      value: "a",
      isOpen: true,
    },
    {
      id: "b",
      value: "b",
      isOpen: true,
      children: [
        {
          id: "c",
          value: "c",
          isOpen: true,
          children: [
            {
              id: "d",
              value: "d",
              isOpen: true,
            },
          ],
        },
        {
          id: "e",
          value: "e",
          isOpen: true,
        },
      ],
    },
  ],
};

const r = data;
const b = r.children![1];
const ba = b.children![0];

describe("TreeView", () => {
  test("Should render elements", async () => {
    render(<TreeView tree={data} />);

    const testElement = async (e: TreeData) => {
      expect(e.value).toBeTruthy();

      const el = await screen.findByText(e.value!.toString());

      waitFor(() => expect(el).toBeVisible());

      e.children?.forEach(testElement);
    };

    await testElement(data);
  });

  test("Should call onClick callback", async () => {
    const user = userEvent.setup();

    const callback = vi.fn((x: TreeData[]) => {
      if (x) return;
    });

    render(<TreeView tree={data} onClick={callback} />);

    const path = [r, b, ba];

    const el = await screen.findByText(ba.value!.toString());

    await user.click(el);

    waitFor(() => expect(callback).toHaveBeenCalledWith(path));
  });

  test("Should hide closed elements", () => {
    const d = produce(data, (draft) => {
      const gotB = draft.children?.find((v) => v.id === b.id);
      expect(gotB).toBeTruthy();

      gotB!.isOpen = false;
    });

    render(<TreeView tree={d} />);

    waitFor(() =>
      expect(screen.findByText(ba.value!.toString())).not.toBeVisible(),
    );
  });
});
