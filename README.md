# aresize
Another Resize Utility for JPEG and PNG
===

Introduction
---
Got tired of having to provide a lot of parameters to tools get a bunch of JPEG/PNG resized for sharing. There's the Batch (B) process mode from IrfanView but on non-Windows platforms there aren't much choices for quick-and-easy batch resizing. Hence, I took some time to come up with this tiny portable program to help me do that.

Compilation
---
Should be extra easy as I hate tool-chain setup as well. 
1. Install golang 1.24+ (1.24.2 was used for development)
2. Git Clone this project 
3. Run ``go build`` to build the binary for your machine, [or you can cross-compile](https://go.dev/wiki/GccgoCrossCompilation)
4. There should be an executable built (aresize.exe for Windows)

Usage
---

    PS ...> aresize --help
    Usage of C:\Users\Frederick\go\bin\aresize.exe:
    -betterResize
            Use CatmullRom instead of BiLinear for resize
    -c int
            Concurrent conversions allowed. This is usually memory bound. (default 4)
    -long int
            The length in pixel of the long side to resize the image (default 2560)
    -p string
            File Glob pattern to search for image files
    -pre string
            Prefix added to the resized image files' name (default "resized_")
    -q int
            JPEG Quality of the output image file. Only used for JPEG (default 100)

Quick Example:
``aresize.exe -p *.jpg``

Some More Explanation
---
By default, it will do the following:
1. Take the Glob pattern (e.g. ``-p *.jpg``) to find out what to process (only supports JPEG and PNG)
2. Resize the image with the longest edge set to 2560px by default, kept aspect ratio and put it under the same directory as source, with a new name (default ``resize_``). And of the same file format. 
3. It will overwrite existing files!

There are some parameters you can play with:
1. Specify the prefix of the newly resized image. 
2. Set the long edge pixel size
2. Specify if you want to use a "better quality" scale algorithm.
3. Level of concurrency (practically just how many goroutines to use for resize). Default to 4 but this is mostly memory bound as a 20MB JPEG can decode to 1GB of size in memory! So choose wisely.

There are some caveats
1. Not the most performant as this is using the built-in ``image`` library from golang. However, it is self-contained being pure-go.
2. EXIF is not kept. There aren't much good options to read and write EXIF on JPEG with pure-go libraries (they either need external tool/libraries in other languages, or just wasn't maintained anymore)

So... just enjoy the simplicity! And probably laugh about my amateur go programming skills (I usually use Python).  
