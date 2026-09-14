import { Avatar, ListBox } from "@heroui/react";
import type { Tenant } from "../../model/model";

/**
 * TenantList properties.
 */
type Props = {
  /** List of tenants. */
  tenants: Tenant[];

  /** Controlled selected ID, leave undefined for uncontrolled behavior. */
  selectedID?: Tenant["ID"];
};

/**
 * Renders a list of tenants.
 */
function TenantList({ tenants, selectedID }: Props) {
  return (
    <ListBox
      aria-label="Tenants"
      selectionMode="single"
      selectedKeys={selectedID && [selectedID]}
    >
      {tenants.map((t) => (
        <ListBox.Item
          key={t.ID}
          id={t.ID}
          textValue={t.Name}
          className="data-selected:bg-neutral-100 rounded-lg"
        >
          <Avatar size="sm" className="rounded-lg">
            <Avatar.Image alt={t.Name} src={t.IconURL} />
            <Avatar.Fallback>{t.Name[0]}</Avatar.Fallback>
          </Avatar>
          <h3>{t.Name}</h3>
          <ListBox.ItemIndicator />
        </ListBox.Item>
      ))}
    </ListBox>
  );
}

export default TenantList;
