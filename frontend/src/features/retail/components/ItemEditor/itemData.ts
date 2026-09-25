import z from "zod";

export const ItemSchema = z.object({
  Name: z.string().nonempty("name is required").nonoptional("name is required"),
  Desc: z.string().optional(),
});

export type ItemData = z.infer<typeof ItemSchema>;
