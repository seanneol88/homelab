#!/usr/bin/env python3
"""pdf2slides — every pdf page becomes a full-slide picture in a .pptx."""
import os, subprocess, sys, tempfile
from pptx import Presentation
from PIL import Image

EMU_PER_IN = 914400
SLIDE_W, SLIDE_H = int(13.333 * EMU_PER_IN), int(7.5 * EMU_PER_IN)

def main(src, dst):
    with tempfile.TemporaryDirectory() as td:
        prefix = os.path.join(td, "page")
        subprocess.run(["pdftoppm", "-png", "-r", "150", src, prefix], check=True)
        pages = sorted(p for p in os.listdir(td) if p.startswith("page"))
        if not pages:
            raise SystemExit("no pages rendered")
        prs = Presentation()
        prs.slide_width, prs.slide_height = SLIDE_W, SLIDE_H
        blank = prs.slide_layouts[6]
        for p in pages:
            img = os.path.join(td, p)
            slide = prs.slides.add_slide(blank)
            iw, ih = Image.open(img).size
            scale = min(SLIDE_W / iw, SLIDE_H / ih)
            w, h = int(iw * scale), int(ih * scale)
            slide.shapes.add_picture(img, (SLIDE_W - w) // 2, (SLIDE_H - h) // 2, width=w, height=h)
        prs.save(dst)

if __name__ == "__main__":
    main(sys.argv[1], sys.argv[2])
