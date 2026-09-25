import {
  Button,
  FieldError,
  Form,
  Input,
  Label,
  TextArea,
  TextField,
  toast,
  Typography,
} from "@heroui/react";
import { useState } from "react";
import ImagePicker from "../../../../components/ImagePicker/ImagePicker";
import { ItemSchema, type ItemData } from "./itemData";
import z from "zod";

/**
 * ItemEditor properties.
 */
type Props = {
  /**
   * Controlled item value.
   */
  value?: ItemData;

  /**
   * Property change callback.
   */
  onChange?: (property: keyof ItemData, value: unknown) => void;

  /**
   * Item create callback.
   */
  onSumbit?: (v: ItemData, image: File | null) => void;

  /**
   * Selected item image, undefined for uncontrolled value.
   */
  selectedImage?: File | null;

  /**
   * On item change callback.
   */
  onSelectedImageChange?: (f: File | null) => void;
};

/**
 * ItemEditor form.
 */
function ItemEditor({
  value,
  onChange,
  onSumbit,
  selectedImage,
  onSelectedImageChange,
}: Props) {
  const [uncontrolledImage, setUncontrolledImage] = useState<File | null>(null);

  const image = selectedImage !== undefined ? selectedImage : uncontrolledImage;

  const handleImageChange = (f: File | null) => {
    if (onSelectedImageChange) {
      onSelectedImageChange(f);
    }

    setUncontrolledImage(f);
  };

  const handleChange = (property: keyof ItemData, value: unknown) => {
    if (onChange) {
      onChange(property, value);
    }
  };

  const handleSumbit = async (data: FormData) => {
    const raw = Object.fromEntries(data.entries());

    const result = await ItemSchema.safeParseAsync(raw);

    if (!result.success) {
      toast(z.prettifyError(result.error));

      return;
    }

    setUncontrolledImage(null);

    if (onSumbit) {
      onSumbit(result.data, image);
    }
  };

  const handleReset = () => {
    setUncontrolledImage(null);

    if (onSelectedImageChange) {
      onSelectedImageChange(null);
    }
  };

  return (
    <div className="w-full h-full p-6">
      <Form action={handleSumbit} className="w-full h-full flex flex-row gap-8">
        <div className="flex-1 h-fit aspect-square rounded-2xl">
          <ImagePicker image={image} onImageChange={handleImageChange} />
        </div>

        <div className="flex-4">
          <div className="w-full h-full flex flex-col">
            <div className="w-full flex-1">
              <Typography type="h3">Create Item</Typography>

              <TextField
                isRequired
                onChange={(v) => handleChange("Name", v)}
                name="Name"
                validate={(v) => {
                  const result = ItemSchema.shape.Name.safeParse(v);

                  if (!result.success) {
                    return result.error.issues.map((e) => e.message).join("\n");
                  }
                }}
              >
                <Label>Name</Label>
                <Input value={value?.Name} />
                <FieldError />
              </TextField>

              <TextField
                onChange={(v) => handleChange("Desc", v)}
                name="Desc"
                validate={(v) => {
                  const result = ItemSchema.shape.Desc.safeParse(v);

                  if (!result.success) {
                    return result.error.issues.map((e) => e.message).join("\n");
                  }
                }}
              >
                <Label>Description</Label>
                <TextArea value={value?.Desc} />
                <FieldError />
              </TextField>
            </div>

            <div className="w-full flex justify-end gap-2">
              <Button type="reset" variant="secondary" onPress={handleReset}>
                Reset
              </Button>
              <Button type="submit" variant="primary">
                Create
              </Button>
            </div>
          </div>
        </div>
      </Form>
    </div>
  );
}

export default ItemEditor;
