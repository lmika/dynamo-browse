package dynamotableview

type columnModel struct {
	m *Model
}

func (cm columnModel) Len() int {
	return len(cm.m.resultSet.Columns()[cm.m.colOffset:]) + 1
}

func (cm columnModel) Header(index int) string {
	if index == 0 {
		return ""
	}

	return cm.m.resultSet.Columns()[cm.m.colOffset+index-1]
}
