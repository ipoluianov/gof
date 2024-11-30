package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation
#cgo LDFLAGS: -framework Cocoa

#import <Foundation/Foundation.h>
#import <Cocoa/Cocoa.h>

@interface AppDelegate : NSObject <NSApplicationDelegate, NSWindowDelegate>
@property (assign) NSWindow *window;
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

    [NSApp activateIgnoringOtherApps:YES];
}

- (void)windowDidResize:(NSNotification *)notification {
    NSRect frame = [self.window frame];
    NSLog(@"Window resized to: %.0fx%.0f", frame.size.width, frame.size.height);
}

- (BOOL)applicationShouldTerminateAfterLastWindowClosed:(NSApplication *)sender {
    return YES;
}

- (void)windowWillClose:(NSNotification *)notification {
    [NSApp terminate:nil];
}
@end

void RunApplication() {
    [NSApplication sharedApplication];
    AppDelegate *delegate = [[AppDelegate alloc] init];
    [NSApp setDelegate:delegate];
    [NSApp run];
}

*/
import "C"
import (
	"image"
	"image/color"
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
	C.RunApplication()
}
