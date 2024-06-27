package ui

import "azdo-dash/ui/section"

func (m *Model) getCurrSection() section.Section {
	section := m.getCurrentViewSections()
	return section
}

func (m *Model) getCurrentViewSections() section.Section {
	return m.prSection
}
