package widget

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/internal/widget"
	"fyne.io/fyne/theme"
)

var _ fyne.Widget = (*menuItem)(nil)

// menuItem is a widget for displaying a fyne.menuItem.
type menuItem struct {
	widget.Base
	Item   *fyne.MenuItem
	Parent *Menu

	child   *Menu
	hovered bool
}

// newMenuItem creates a new menuItem.
func newMenuItem(item *fyne.MenuItem, parent *Menu) *menuItem {
	return &menuItem{Item: item, Parent: parent}
}

// newMenuItemSeparator creates a separator meant to separate MenuItems.
func newMenuItemSeparator() fyne.CanvasObject {
	s := canvas.NewRectangle(theme.DisabledTextColor())
	s.SetMinSize(fyne.NewSize(1, 2))
	return s
}

// CreateRenderer returns a new renderer for the menu item.
// Implements: fyne.Widget
func (i *menuItem) CreateRenderer() fyne.WidgetRenderer {
	text := canvas.NewText(i.Item.Label, theme.TextColor())
	objects := []fyne.CanvasObject{text}
	var icon *canvas.Image
	if i.Item.ChildMenu != nil {
		icon = canvas.NewImageFromResource(theme.MenuExpandIcon())
		objects = append(objects, icon)
		if i.child == nil {
			child := NewMenu(i.Item.ChildMenu)
			child.Hide()
			child.DismissAction = i.Parent.Dismiss
			i.child = child
		}
		objects = append(objects, i.child)
	}
	return &menuItemRenderer{
		BaseRenderer: widget.NewBaseRenderer(objects),
		i:            i,
		icon:         icon,
		text:         text,
	}
}

// Hide hides the menu item.
// Implements: fyne.Widget
func (i *menuItem) Hide() {
	widget.HideWidget(&i.Base, i)
}

// MinSize returns the minimal size of the menu item.
// Implements: fyne.Widget
func (i *menuItem) MinSize() fyne.Size {
	return widget.MinSizeOf(i)
}

// MouseIn changes the item to be hovered and shows the submenu if the item has one.
// The submenu of any sibling of the item will be hidden.
// Implements: desktop.Hoverable
func (i *menuItem) MouseIn(*desktop.MouseEvent) {
	i.hovered = true
	i.activateChild()
	i.Refresh()
}

// MouseMoved does nothing.
// Implements: desktop.Hoverable
func (i *menuItem) MouseMoved(*desktop.MouseEvent) {
}

// MouseOut changes the item to not be hovered but has no effect on the visibility of the item's submenu.
// Implements: desktop.Hoverable
func (i *menuItem) MouseOut() {
	i.hovered = false
	i.Refresh()
}

// Refresh triggers a redraw of the menu item.
// Implements: fyne.Widget
func (i *menuItem) Refresh() {
	widget.RefreshWidget(i)
}

// Resize changes the size of the menu item.
// Implements: fyne.Widget
func (i *menuItem) Resize(size fyne.Size) {
	widget.ResizeWidget(&i.Base, i, size)
	if i.child != nil {
		i.updateChildPosition()
	}
}

// Show makes the menu item visible.
// Implements: fyne.Widget
func (i *menuItem) Show() {
	widget.ShowWidget(&i.Base, i)
}

// Tapped performs the action of the item and dismisses the menu.
// It does nothing if the item doesn’t have an action.
// Implements: fyne.Tappable
func (i *menuItem) Tapped(*fyne.PointEvent) {
	if i.Item.Action == nil {
		if fyne.CurrentDevice().IsMobile() {
			i.activateChild()
			i.Refresh()
		}
		return
	}

	i.Item.Action()
	i.Parent.Dismiss()
}

func (i *menuItem) activateChild() {
	if i.child != nil {
		i.child.DeactivateChild()
	}
	if i.Parent.activeChild == i.child {
		return
	}

	i.Parent.DeactivateChild()
	if i.child != nil {
		if i.child.Size().IsZero() {
			i.child.Resize(i.child.MinSize())
			i.updateChildPosition()
		}
		i.Parent.activeChild = i.child
		i.child.Show()
	}
}

func (i *menuItem) updateChildPosition() {
	itemSize := i.Size()
	cp := fyne.NewPos(itemSize.Width, -theme.Padding())
	d := fyne.CurrentApp().Driver()
	c := d.CanvasForObject(i)
	if c != nil {
		absPos := d.AbsolutePositionForObject(i)
		childSize := i.child.Size()
		if absPos.X+itemSize.Width+childSize.Width > c.Size().Width {
			if absPos.X-childSize.Width >= 0 {
				cp.X = -childSize.Width
			} else {
				cp.X = c.Size().Width - absPos.X - childSize.Width
			}
		}
		if absPos.Y+childSize.Height-theme.Padding() > c.Size().Height {
			cp.Y = c.Size().Height - absPos.Y - childSize.Height
		}
	}
	i.child.Move(cp)
}

type menuItemRenderer struct {
	widget.BaseRenderer
	i                *menuItem
	icon             *canvas.Image
	lastThemePadding int
	minSize          fyne.Size
	text             *canvas.Text
}

func (r *menuItemRenderer) BackgroundColor() color.Color {
	if !fyne.CurrentDevice().IsMobile() && (r.i.hovered || (r.i.child != nil && r.i.child.Visible())) {
		return theme.HoverColor()
	}

	return color.Transparent
}

func (r *menuItemRenderer) Layout(size fyne.Size) {
	padding := r.itemPadding()

	r.text.TextSize = theme.TextSize()
	r.text.Color = theme.TextColor()
	r.text.Resize(r.text.MinSize())
	r.text.Move(fyne.NewPos(padding.Width/2, padding.Height/2))

	if r.icon != nil {
		r.icon.Resize(fyne.NewSize(theme.IconInlineSize(), theme.IconInlineSize()))
		r.icon.Move(fyne.NewPos(size.Width-theme.IconInlineSize(), (size.Height-theme.IconInlineSize())/2))
	}
}

func (r *menuItemRenderer) MinSize() fyne.Size {
	if r.minSizeUnchanged() {
		return r.minSize
	}

	minSize := r.text.MinSize().Add(r.itemPadding())
	if r.icon != nil {
		minSize = minSize.Add(fyne.NewSize(theme.IconInlineSize(), 0))
	}
	r.minSize = minSize
	return r.minSize
}

func (r *menuItemRenderer) Refresh() {
	canvas.Refresh(r.i)
}

func (r *menuItemRenderer) minSizeUnchanged() bool {
	return !r.minSize.IsZero() &&
		r.text.TextSize == theme.TextSize() &&
		(r.icon == nil || r.icon.Size().Width == theme.IconInlineSize()) &&
		r.lastThemePadding == theme.Padding()
}

func (r *menuItemRenderer) itemPadding() fyne.Size {
	return fyne.NewSize(theme.Padding()*4, theme.Padding()*2)
}
