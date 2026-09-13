import { render, screen, waitFor } from "@testing-library/react";
import SidebarLayout from "./SidebarLayout";

describe("SidebarLayout", () => {
  test("Should render main content", async () => {
    const expected = "expected";

    render(
      <SidebarLayout>
        <h1 data-testid={expected}>Hello world</h1>
      </SidebarLayout>,
    );

    const el = await screen.findByTestId(expected);

    expect(el).toBeVisible();
  });

  test("Should render sidebar when open", async () => {
    const expected = "expected";

    render(
      <SidebarLayout sidebar={<h2 data-testid={expected}>abc</h2>} isOpen>
        <h1>Hello world</h1>
      </SidebarLayout>,
    );

    const el = await screen.findByTestId(expected);

    expect(el).toBeVisible();
  });

  test("Should not render sidebar when closed", async () => {
    const expected = "expected";

    render(
      <SidebarLayout
        sidebar={<h2 data-testid={expected}>abc</h2>}
        isOpen={false}
      >
        <h1>Hello world</h1>
      </SidebarLayout>,
    );

    const el = await screen.findByTestId(expected);

    waitFor(() => {
      expect(el).not.toBeVisible();
    });
  });
});
