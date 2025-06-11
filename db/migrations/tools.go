package migrations

import (
	"fmt"
	"slices"

	"github.com/pocketbase/pocketbase/core"
)

func ID(name string) string {
	return fmt.Sprintf("__%s__", name)
}

func addUpdatedFields(fields *core.FieldsList) {
	fields.Add(
		&core.AutodateField{
			Id: ID("created"), Name: "created", System: true,
			OnCreate: true,
		},
		&core.AutodateField{
			Id: ID("updated"), Name: "updated", System: true,
			OnCreate: true,
			OnUpdate: true,
		},
	)
}

func getFieldIndex(m *core.Collection, name string) int {
	return slices.Index(m.Fields.FieldNames(), name)
}
