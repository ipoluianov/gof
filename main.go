package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation
#cgo LDFLAGS: -framework Cocoa

#import <Foundation/Foundation.h>
#import <Cocoa/Cocoa.h>

// Кастомный класс NSView для отрисовки изображения
@interface CustomImageView : NSView
@property (assign) CGImageRef image;
@end

@implementation CustomImageView

- (void)drawRect:(NSRect)dirtyRect {
    [super drawRect:dirtyRect];

    if (self.image) {
        CGContextRef context = [[NSGraphicsContext currentContext] CGContext];
        CGRect rect = NSRectToCGRect([self bounds]);
        CGContextDrawImage(context, rect, self.image);
    }
}

- (void)setImage:(CGImageRef)newImage {
    if (_image) {
        CGImageRelease(_image); // Освобождаем старое изображение
    }
    _image = CGImageRetain(newImage); // Сохраняем новое изображение
    [self setNeedsDisplay:YES]; // Запрашиваем перерисовку
}

- (void)dealloc {
    if (_image) {
        CGImageRelease(_image); // Освобождаем изображение при удалении
    }
    [super dealloc];
}

@end

@interface AppDelegate : NSObject <NSApplicationDelegate, NSWindowDelegate>
@property (assign) NSWindow *window;
@property (assign) CustomImageView *imageView; // Используем кастомный NSView
@end

@implementation AppDelegate
- (void)applicationDidFinishLaunching:(NSNotification *)notification {
    [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];

    NSRect frame = NSMakeRect(100, 100, 800, 600);
    NSUInteger style = NSWindowStyleMaskTitled | NSWindowStyleMaskResizable | NSWindowStyleMaskClosable;
    self.window = [[NSWindow alloc] initWithContentRect:frame
                                               styleMask:style
                                                 backing:NSBackingStoreBuffered
                                                   defer:NO];
    [self.window setTitle:@"MacOS Window"];
    [self.window setDelegate:self];
    [self.window makeKeyAndOrderFront:nil];

	 // Создаем CustomImageView
    self.imageView = [[CustomImageView alloc] initWithFrame:[[self.window contentView] bounds]];
    [[self.window contentView] addSubview:self.imageView];

    [NSApp activateIgnoringOtherApps:YES];
}

- (void)setImageWithBytes:(void *)bytes width:(int)width height:(int)height {
    NSLog(@"Setting image with width: %d, height: %d", width, height);
    CGColorSpaceRef colorSpace = CGColorSpaceCreateDeviceRGB();
    CGContextRef context = CGBitmapContextCreate(
        bytes,
        width,
        height,
        8, // 8 бит на компонент
        width * 4, // 4 байта на пиксель (RGBA)
        colorSpace,
        kCGImageAlphaPremultipliedLast | kCGBitmapByteOrderDefault
    );

    if (!context) {
        NSLog(@"Failed to create CGContext");
    }

    CGImageRef image = CGBitmapContextCreateImage(context);
    if (!image) {
        NSLog(@"Failed to create CGImage");
    } else {
        NSLog(@"Image created successfully");
    }

    [self.imageView setImage:image];

    CGImageRelease(image);
    CGContextRelease(context);
    CGColorSpaceRelease(colorSpace);
}

- (void)windowDidResize:(NSNotification *)notification {
    NSRect frame = [self.window frame];
    [self.imageView setFrame:[[self.window contentView] bounds]]; // Подгоняем под размер окна
    [self.imageView setNeedsDisplay:YES]; // Перерисовываем
}

- (BOOL)applicationShouldTerminateAfterLastWindowClosed:(NSApplication *)sender {
    return YES;
}

- (void)windowWillClose:(NSNotification *)notification {
    [NSApp terminate:nil];
}


@end

AppDelegate *delegate;

void RunApplication() {
    [NSApplication sharedApplication];
    delegate = [[AppDelegate alloc] init];
    [NSApp setDelegate:delegate];
    [NSApp run];
}

void SetImage(void *bytes, int width, int height) {
    [delegate setImageWithBytes:bytes width:width height:height];
}

*/
import "C"
import (
	"image"
	"image/color"
	"unsafe"
)

func generateRGBA(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(50),
				G: uint8(200),
				B: uint8(200),
				A: 255,
			})
		}
	}
	return img.Pix
}

func main() {
	const width, height = 800, 600
	rgbaBytes := generateRGBA(width, height)
	C.SetImage(unsafe.Pointer(&rgbaBytes[0]), C.int(width), C.int(height))
	C.RunApplication()
}
