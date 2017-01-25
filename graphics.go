package pixel

import (
	"fmt"
	"image/color"
)

// TrianglesData specifies a list of Triangles vertices with three common properties: Position,
// Color and Texture.
type TrianglesData []struct {
	Position Vec
	Color    NRGBA
	Texture  Vec
}

// Len returns the number of vertices in TrianglesData.
func (td *TrianglesData) Len() int {
	return len(*td)
}

// Draw is unimplemented for TrianglesData and panics.
func (td *TrianglesData) Draw() {
	panic(fmt.Errorf("%T.Draw: invalid operation", td))
}

func (td *TrianglesData) resize(len int) {
	if len > td.Len() {
		needAppend := len - td.Len()
		for i := 0; i < needAppend; i++ {
			*td = append(*td, struct {
				Position Vec
				Color    NRGBA
				Texture  Vec
			}{V(0, 0), NRGBA{1, 1, 1, 1}, V(-1, -1)})
		}
	}
	if len < td.Len() {
		*td = (*td)[:len]
	}
}

func (td *TrianglesData) updateData(offset int, t Triangles) {
	// fast path optimization
	if t, ok := t.(*TrianglesData); ok {
		copy((*td)[offset:], *t)
		return
	}

	// slow path manual copy
	if t, ok := t.(TrianglesPosition); ok {
		for i := offset; i < len(*td); i++ {
			(*td)[i].Position = t.Position(i)
		}
	}
	if t, ok := t.(TrianglesColor); ok {
		for i := offset; i < len(*td); i++ {
			(*td)[i].Color = t.Color(i)
		}
	}
	if t, ok := t.(TrianglesTexture); ok {
		for i := offset; i < len(*td); i++ {
			(*td)[i].Texture = t.Texture(i)
		}
	}
}

// Update copies vertex properties from the supplied Triangles into this TrianglesData.
//
// TrianglesPosition, TrianglesColor and TrianglesTexture are supported.
func (td *TrianglesData) Update(t Triangles) {
	td.resize(t.Len())
	td.updateData(0, t)
}

// Append adds supplied Triangles to the end of the TrianglesData.
func (td *TrianglesData) Append(t Triangles) {
	td.resize(td.Len() + t.Len())
	td.updateData(td.Len()-t.Len(), t)
}

// Copy returns an exact independent copy of this TrianglesData.
func (td *TrianglesData) Copy() Triangles {
	copyTd := make(TrianglesData, td.Len())
	copyTd.Update(td)
	return &copyTd
}

// Position returns the position property of i-th vertex.
func (td *TrianglesData) Position(i int) Vec {
	return (*td)[i].Position
}

// Color returns the color property of i-th vertex.
func (td *TrianglesData) Color(i int) NRGBA {
	return (*td)[i].Color
}

// Texture returns the texture property of i-th vertex.
func (td *TrianglesData) Texture(i int) Vec {
	return (*td)[i].Texture
}

// TrianglesDrawer is a helper type that wraps Triangles and turns them into a Drawer.
//
// It does so by creating a separate Triangles instance for each Target. The instances are
// correctly updated alongside the wrapped Triangles.
type TrianglesDrawer struct {
	Triangles

	tris  map[Target]Triangles
	dirty bool
}

func (td *TrianglesDrawer) flush() {
	if !td.dirty {
		return
	}
	td.dirty = false

	for _, t := range td.tris {
		t.Update(td.Triangles)
	}
}

// Draw draws the wrapped Triangles onto the provided Target.
func (td *TrianglesDrawer) Draw(target Target) {
	if td.tris == nil {
		td.tris = make(map[Target]Triangles)
	}

	td.flush()

	tri := td.tris[target]
	if tri == nil {
		tri = target.MakeTriangles(td.Triangles)
		td.tris[target] = tri
	}
	tri.Draw()
}

// Update updates the wrapped Triangles with the supplied Triangles.
//
// Call only this method to update the wrapped Triangles, otherwise the TrianglesDrawer will not
// work correctly.
func (td *TrianglesDrawer) Update(t Triangles) {
	td.dirty = true
	td.Triangles.Update(t)
}

// Append appends the supplied Triangles to the wrapped Triangles.
//
// Call only this method to append to the wrapped Triangles, otherwise the TrianglesDrawer will not
// work correctly.
func (td *TrianglesDrawer) Append(t Triangles) {
	td.dirty = true
	td.Triangles.Append(t)
}

// Sprite is a picture that can be drawn onto a Target. To change the position/rotation/scale of
// the Sprite, use Target's SetTransform method.
type Sprite struct {
	td   TrianglesDrawer
	data *TrianglesData
	pic  *Picture
}

// NewSprite creates a Sprite with the supplied Picture. The dimensions of the returned Sprite match
// the dimensions of the Picture.
func NewSprite(pic *Picture) *Sprite {
	data := TrianglesData{
		{Position: V(0, 0), Color: NRGBA{1, 1, 1, 1}, Texture: V(0, 0)},
		{Position: V(0, 0), Color: NRGBA{1, 1, 1, 1}, Texture: V(1, 0)},
		{Position: V(0, 0), Color: NRGBA{1, 1, 1, 1}, Texture: V(1, 1)},
		{Position: V(0, 0), Color: NRGBA{1, 1, 1, 1}, Texture: V(0, 0)},
		{Position: V(0, 0), Color: NRGBA{1, 1, 1, 1}, Texture: V(1, 1)},
		{Position: V(0, 0), Color: NRGBA{1, 1, 1, 1}, Texture: V(0, 1)},
	}
	s := &Sprite{
		td:   TrianglesDrawer{Triangles: &data},
		data: &data,
	}
	s.SetPicture(pic)
	return s
}

// SetPicture changes the Picture of the Sprite and resizes it accordingly.
func (s *Sprite) SetPicture(pic *Picture) {
	w, h := pic.Bounds().Size.XY()
	(*s.data)[0].Position = V(0, 0)
	(*s.data)[2].Position = V(w, h)
	(*s.data)[1].Position = V(w, 0)
	(*s.data)[3].Position = V(0, 0)
	(*s.data)[4].Position = V(w, h)
	(*s.data)[5].Position = V(0, h)
	s.pic = pic
}

// Picture returns the current Picture of the Sprite.
func (s *Sprite) Picture() *Picture {
	return s.pic
}

// Draw draws the Sprite onto the provided Target.
func (s *Sprite) Draw(t Target) {
	t.SetPicture(s.pic)
	s.td.Draw(t)
}

// Polygon is a convex polygon shape filled with a single color.
type Polygon struct {
	td   TrianglesDrawer
	data *TrianglesData
	col  NRGBA
}

// NewPolygon creates a Polygon with specified color and points. Points can be in clock-wise or
// counter-clock-wise order, it doesn't matter. They should however form a convex polygon.
func NewPolygon(c color.Color, points ...Vec) *Polygon {
	data := make(TrianglesData, len(points))
	p := &Polygon{
		td:   TrianglesDrawer{Triangles: &data},
		data: &data,
	}
	p.SetColor(c)
	p.SetPoints(points...)
	return p
}

// SetColor changes the color of the Polygon.
//
// If the Polygon is very large, this method might end up being too expensive. Consider using
// a color mask on a Target, in such a case.
func (p *Polygon) SetColor(c color.Color) {
	p.col = NRGBAModel.Convert(c).(NRGBA)
	for i := range *p.data {
		(*p.data)[i].Color = p.col
	}
	// dirty stuff, need to update manually
	p.td.dirty = true
}

// Color returns the current color of the Polygon.
func (p *Polygon) Color() NRGBA {
	return p.col
}

// SetPoints sets the points of the Polygon. The number of points might differ from the original
// count.
//
// This method is more effective, than creating a new Polygon with the given points.
func (p *Polygon) SetPoints(points ...Vec) {
	p.data.resize(len(points))
	for i, pt := range points {
		(*p.data)[i].Position = pt
		(*p.data)[i].Color = p.col
		(*p.data)[i].Texture = V(-1, -1)
	}
	// dirty stuff
	p.td.dirty = true
}

// Points returns a slice of points of the Polygon in the order they where supplied.
func (p *Polygon) Points() []Vec {
	points := make([]Vec, p.data.Len())
	for i := range *p.data {
		points[i] = (*p.data)[i].Position
	}
	return points
}

// Draw draws the Polygon onto the Target.
func (p *Polygon) Draw(t Target) {
	t.SetPicture(nil)
	p.td.Draw(t)
}
