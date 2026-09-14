/**
 * Tenant represents a tenant that can own retails and have users.
 */
type Tenant = {
  ID: string;
  Name: string;
  IconURL?: string;
  Retails: Retail[];
};

type Retail = {
  ID: string;
  Name: string;
  Storage: Place;
};

type Place = {
  ID: string;
  Name: string;
  Items?: StockItem[];
  Children?: Place[];
};

type StockItem = {
  ID: string;
  Stock?: number;
  Data: Item;
};

type Item = {
  ID: string;
  Name: string;
  Desc?: string;
};

export type { Tenant, Retail, Place, StockItem, Item };
