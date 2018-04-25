package watermark

import (
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/yudeguang/haserr"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"os"
	"strings"
)

var (
	defaultFontGrayscaleColor  = uint8(0)    //黑色
	defaultFontSize            = float64(14) //默认为14
	defaultDpi                 = float64(72)
	defaultFontAdress          = `C:\Windows\Fonts\simhei.ttf` //黑体常规
	defaultWatermarkStartPoint = image.Pt(0, 0)                //默认从图片的左上角开始写入水印
	defaultResizeWith          = 0                             //默认0,宽度不调整
	defaultResizeHeight        = 0                             //默认0,高度不调整
)

//签名结构体
type Watermark struct {
	FontGrayscaleColor  uint8          //字体灰度颜色: 0 黑色 0xffff 白色
	FontSize            float64        //字体大小
	Dpi                 float64        //分辨率
	FontAdress          string         //字体在电脑上的物理路径，注意，黑体，楷体，宋体等少数中文字体无版权，千万不要用微软雅黑
	font                *truetype.Font //字体
	WatermarkStartPoint image.Point    //在图片上签名开始位置（点）
	ResizeWith          int            //图片调整后的新宽度 0表示不修改
	ResizeHeight        int            //图片调整后的新高度 0表示不修改
}

//全部默认实例化
func NewDefaultWatermark() (*Watermark, error) {
	return NewWatermark(defaultFontAdress,
		defaultFontGrayscaleColor,
		defaultFontSize,
		defaultDpi,
		defaultWatermarkStartPoint,
		defaultResizeWith,
		defaultResizeHeight)
}

//自定义签名
func NewWatermark(fontAdress string, grayscaleColor uint8, fontSize, dpi float64, WatermarkStartPoint image.Point, resizeWith, resizeHeight int) (*Watermark, error) {
	font, err := loadingFont(fontAdress)
	if err != nil {
		return nil, err
	}
	return &Watermark{
		FontGrayscaleColor:  grayscaleColor,
		FontSize:            fontSize,
		font:                font,
		Dpi:                 dpi,
		WatermarkStartPoint: WatermarkStartPoint,
		ResizeWith:          resizeWith,
		ResizeHeight:        resizeHeight,
	}, nil
}

//重新设置水印在原始图片开始的位置
func (this *Watermark) SetWatermarkStartPoint(X, Y int) {
	this.WatermarkStartPoint.X, this.WatermarkStartPoint.Y = X, Y
}

//加载字体
func loadingFont(fontPath string) (*truetype.Font, error) {
	fontBytes, err := ioutil.ReadFile(fontPath)
	if err != nil {
		return nil, err
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, err
	}
	return font, nil
}

//在图片上加载文字水印
func (this *Watermark) Watermark(srcFile, dstFile string, text string) error {
	output, err := os.OpenFile(dstFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	haserr.Fatal(err)
	defer output.Close()
	origin, err := imaging.Open(srcFile)
	if err != nil {
		return fmt.Errorf("image decode error(%v)", err)
	}
	//重新设定图片最大宽度
	if this.ResizeWith == 0 {
		this.ResizeWith = origin.Bounds().Max.X
	}
	origin = imaging.Resize(origin, this.ResizeWith, this.ResizeHeight, imaging.Lanczos)

	dst := image.NewNRGBA(origin.Bounds())
	draw.Draw(dst, dst.Bounds(), origin, image.ZP, draw.Src)
	mask, err := this.drawStringImage(text, origin.Bounds().Max.X, int(defaultFontSize)+1)
	if err != nil {
		return fmt.Errorf("image decode error(%v)", err)
	}
	draw.Draw(dst, mask.Bounds().Add(this.WatermarkStartPoint), mask, image.ZP, draw.Over)
	return imaging.Save(dst, dstFile)
}

// 画一个带有text的透明图片，宽度为待打水印图片宽度，高度为字体大小加1
func (this *Watermark) drawStringImage(text string, width, height int) (image.Image, error) {
	fg, bg := image.NewUniform(color.Gray{this.FontGrayscaleColor}), image.Transparent
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)

	c := freetype.NewContext()
	c.SetDPI(this.Dpi)
	c.SetFont(this.font)
	c.SetFontSize(this.FontSize)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fg)
	// Draw the text.
	pt := freetype.Pt(10, 10+int(c.PointToFixed(12)>>8))
	for _, s := range strings.Split(text, "\r\n") {
		_, err := c.DrawString(s, pt)
		if err != nil {
			err := fmt.Errorf("c.DrawString(%s) error(%v)", s, err)
			return nil, err
		}
		pt.Y += c.PointToFixed(12 * 1.5)
	}
	return rgba, nil
}
