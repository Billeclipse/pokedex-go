from __future__ import annotations

import json
from pathlib import Path

from reportlab.lib import colors
from reportlab.lib.pagesizes import LETTER
from reportlab.lib.styles import ParagraphStyle, getSampleStyleSheet
from reportlab.lib.units import inch
from reportlab.platypus import (
    BaseDocTemplate,
    Frame,
    PageTemplate,
    Paragraph,
    Spacer,
)


ROOT = Path(__file__).resolve().parents[1]
NOTES_PATH = ROOT / "docs" / "lesson_notes.json"
OUTPUT_PATH = ROOT / "output" / "pdf" / "pokedex-go-lessons.pdf"


def draw_page(canvas, doc):
    canvas.saveState()
    width, _ = LETTER
    canvas.setStrokeColor(colors.HexColor("#d9e2ec"))
    canvas.line(inch * 0.75, inch * 0.65, width - inch * 0.75, inch * 0.65)
    canvas.setFont("Helvetica", 9)
    canvas.setFillColor(colors.HexColor("#52616f"))
    canvas.drawString(inch * 0.75, inch * 0.42, "Pokedex Go Lessons")
    canvas.drawRightString(width - inch * 0.75, inch * 0.42, f"Page {doc.page}")
    canvas.restoreState()


def build_pdf():
    with NOTES_PATH.open("r", encoding="utf-8") as source:
        notes = json.load(source)

    OUTPUT_PATH.parent.mkdir(parents=True, exist_ok=True)

    styles = getSampleStyleSheet()
    title_style = ParagraphStyle(
        "ProjectTitle",
        parent=styles["Title"],
        fontName="Helvetica-Bold",
        fontSize=24,
        leading=30,
        textColor=colors.HexColor("#102a43"),
        spaceAfter=10,
    )
    subtitle_style = ParagraphStyle(
        "Subtitle",
        parent=styles["Normal"],
        fontName="Helvetica",
        fontSize=10,
        leading=14,
        textColor=colors.HexColor("#52616f"),
        spaceAfter=24,
    )
    lesson_title_style = ParagraphStyle(
        "LessonTitle",
        parent=styles["Heading2"],
        fontName="Helvetica-Bold",
        fontSize=15,
        leading=20,
        textColor=colors.HexColor("#243b53"),
        spaceBefore=10,
        spaceAfter=8,
    )
    body_style = ParagraphStyle(
        "LessonBody",
        parent=styles["BodyText"],
        fontName="Helvetica",
        fontSize=10.5,
        leading=15,
        textColor=colors.HexColor("#1f2933"),
        spaceAfter=12,
    )

    doc = BaseDocTemplate(
        str(OUTPUT_PATH),
        pagesize=LETTER,
        leftMargin=inch * 0.75,
        rightMargin=inch * 0.75,
        topMargin=inch * 0.85,
        bottomMargin=inch * 0.85,
        title=notes.get("title", "Pokedex Go Lessons"),
        author="Codex lesson documentation agent",
    )
    frame = Frame(
        doc.leftMargin,
        doc.bottomMargin,
        doc.width,
        doc.height,
        id="normal",
    )
    doc.addPageTemplates([PageTemplate(id="lessons", frames=[frame], onPage=draw_page)])

    story = [
        Paragraph(notes.get("title", "Pokedex Go Lessons"), title_style),
        Paragraph(
            "A running record of the project lessons and the implementation milestone completed in each step.",
            subtitle_style,
        ),
    ]

    for lesson in notes["lessons"]:
        story.append(
            Paragraph(
                f"Lesson {lesson['number']}: {lesson['title']}",
                lesson_title_style,
            )
        )
        story.append(Paragraph(lesson["paragraph"], body_style))
        story.append(Spacer(1, 6))

    doc.build(story)


if __name__ == "__main__":
    build_pdf()
