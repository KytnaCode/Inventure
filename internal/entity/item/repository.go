package item

// Data contains fields necessary to create an item with [Repository].
type Data struct {
	// Name is item's name.
	Name string `validate:"required,max=80,resourcename"`

	// Desc is item's description.
	Desc string `validate:"max=65565,max=0|resourcename"`

	// Stock is item's stock, must be positive.
	Stock int `validate:"gte=0"`

	// Attrs are optional per-item custom attributes.
	Attrs map[string]any `validate:"dive,keys,required,min=1,max=80,endkeys"`
}
