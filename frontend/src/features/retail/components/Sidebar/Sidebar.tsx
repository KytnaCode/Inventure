import { CloseButton, Separator } from "@heroui/react";
import type { Tenant } from "../../model/model";
import RetailTree from "../RetailTree/RetailTree";
import TenantList from "../TenantList/TenantList";

type Props = {
  /** User's tenants. */
  tenants: Tenant[];

  /** Current selected tenant. */
  selectedTenant: Tenant;

  /** onClose handler. */
  onClose?: () => void;
};

/**
 * Sidebar is app's main sidebar.
 */
function Sidebar({ tenants, selectedTenant, onClose }: Props) {
  const handlePress = () => {
    if (onClose) {
      onClose();
    }
  };

  return (
    <div className="w-full h-full p-4 flex flex-col gap-8 overflow-scroll pb-8 ">
      (onClose &&
      <div className="w-full flex flex-row justify-end items-center">
        <CloseButton onPress={handlePress} />
      </div>
      )
      <TenantList tenants={tenants} />
      <Separator />
      <div className="pl-8">
        <RetailTree retails={selectedTenant.Retails} />
      </div>
    </div>
  );
}

export default Sidebar;
