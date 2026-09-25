import { useMemo, useRef, useState, type ChangeEvent } from "react";

/**
 * ImagePicker properties.
 */
type Props = {
  /**
   * Selected image file, undefined for uncontrolled component.
   */
  image?: File | null;

  /**
   * Callback called on image change.
   */
  onImageChange?: (v: File | null) => void;
};

/**
 * Image picker input with preview.
 */
function ImagePicker({ image, onImageChange }: Props) {
  const [uncontrolledImage, setUncontrolledImage] = useState<File | null>(null);
  const ref = useRef<HTMLInputElement>(null);

  const imageFile = image !== undefined ? image : uncontrolledImage;

  const previewUrl = useMemo(
    () => (imageFile ? URL.createObjectURL(imageFile) : null),
    [imageFile],
  );

  const handleButtonClick = () => {
    ref.current?.click();
  };

  const handleFileChange = (e: ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length == 1) {
      const file = e.target.files[0];

      if (!image) {
        setUncontrolledImage(e.target.files[0]);
      }

      if (onImageChange) {
        onImageChange(file);
      }
    }
  };

  return (
    <div className="w-full h-full rounded-2xl overflow-hidden bg-neutral-100 opacity-100 hover:opacity-90">
      <input
        type="file"
        accept="image/*"
        className="hidden"
        onChange={handleFileChange}
        ref={ref}
      />
      <img
        className="w-full h-full object-cover"
        src={previewUrl!}
        onClick={handleButtonClick}
      />
    </div>
  );
}

export default ImagePicker;
