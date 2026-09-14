import { render, screen, waitFor } from "@testing-library/react";
import type { Tenant } from "../../model/model";
import TenantList from "./TenantList";

describe("TenantList", () => {
  test("Should render single tenant", async () => {
    const tenants: Tenant[] = [
      {
        ID: "real-id",
        Name: "Arya's company",
        IconURL:
          "https://i.pinimg.com/236x/67/c0/21/67c0219ba67946108720d31d6c5bb7f7.jpg",
      },
    ];

    render(<TenantList tenants={tenants} />);

    const el = await screen.findByText(tenants[0].Name);

    waitFor(() => expect(el).toBeVisible());
  });

  test("Should render tenant list", async () => {
    const tenants: Tenant[] = [
      {
        ID: "real-id",
        Name: "Arya's company",
        IconURL:
          "https://i.pinimg.com/236x/67/c0/21/67c0219ba67946108720d31d6c5bb7f7.jpg",
      },
      {
        ID: "another-real-id",
        Name: "Ana's company",
        IconURL:
          "https://i.pinimg.com/236x/67/c0/21/67c0219ba67946108720d31d6c5bb7f7.jpg",
      },
      {
        ID: "one-more-real-id",
        Name: "Alex's company",
        IconURL:
          "https://i.pinimg.com/236x/67/c0/21/67c0219ba67946108720d31d6c5bb7f7.jpg",
      },
    ];

    render(<TenantList tenants={tenants} />);

    tenants.forEach(async (v) => {
      const el = await screen.findByText(v.Name);

      waitFor(() => expect(el).toBeVisible());
    });
  });

  test("Should select controlled", async () => {
    const tenants: Tenant[] = [
      {
        ID: "real-id",
        Name: "Arya's company",
        IconURL:
          "https://i.pinimg.com/236x/67/c0/21/67c0219ba67946108720d31d6c5bb7f7.jpg",
      },
      {
        ID: "another-real-id",
        Name: "Ana's company",
        IconURL:
          "https://i.pinimg.com/236x/67/c0/21/67c0219ba67946108720d31d6c5bb7f7.jpg",
      },
      {
        ID: "one-more-real-id",
        Name: "Alex's company",
        IconURL:
          "https://i.pinimg.com/236x/67/c0/21/67c0219ba67946108720d31d6c5bb7f7.jpg",
      },
    ];

    const t = tenants[1];

    render(<TenantList tenants={tenants} selectedID={t.ID} />);

    const el = await screen.findByText(t.Name);

    expect(el.parentElement?.getAttribute("data-selected")).toBeTruthy();
  });
});
