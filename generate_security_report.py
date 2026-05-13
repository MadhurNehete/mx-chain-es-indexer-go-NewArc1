from fpdf import FPDF
import os

class AuditPDF(FPDF):
    def __init__(self):
        super().__init__()
        self.set_margins(20, 20, 20)
        self.set_auto_page_break(auto=True, margin=20)

    def header(self):
        pass

    def footer(self):
        self.set_y(-15)
        self.set_font('helvetica', 'I', 8)
        self.cell(0, 10, f'Page {self.page_no()}', 0, 0, 'C')

    def add_horizontal_line(self):
        self.ln(2)
        self.set_draw_color(0, 0, 0)
        self.set_line_width(0.5)
        self.line(self.get_x(), self.get_y(), self.w - self.get_x(), self.get_y())
        self.ln(5)

    def draw_matrix_row(self, cells, is_header=False):
        if is_header:
            self.set_fill_color(242, 242, 242)
            self.set_font('helvetica', 'B', 10)
        else:
            self.set_fill_color(255, 255, 255)
            self.set_font('helvetica', '', 9)

        # Approximate column widths: ID(25), Finding(100), Severity(25), Verdict(25)
        widths = [20, 100, 25, 25]
        h = 8
        for i, text in enumerate(cells):
            self.cell(widths[i], h, text, border=1, fill=True)
        self.ln(h)

    def draw_alert_block(self, text):
        self.set_font('helvetica', 'I', 10)
        self.set_fill_color(0, 0, 0)
        curr_y = self.get_y()
        # Draw thick left bar
        self.rect(self.get_x(), curr_y, 2, 10, style='F')
        self.set_x(self.get_x() + 5)
        self.write(5, text + '\n')
        self.ln(5)

    def draw_evidence_block(self, code_lines):
        self.set_fill_color(249, 249, 249)
        self.set_font('courier', '', 8)
        # Calculate height needed
        height = len(code_lines) * 4 + 4
        self.rect(self.get_x(), self.get_y(), self.w - 2 * self.get_x(), height, style='F')
        self.set_x(self.get_x() + 2)
        start_y = self.get_y() + 2
        self.set_y(start_y)
        for line in code_lines:
            self.write(4, line + '\n')
            self.set_x(self.get_x() + 2)
        self.ln(5)

def generate_report():
    md_file = "SECURITY_AUDIT_REPORT_Combined.md"
    pdf_file = "SECURITY_AUDIT_REPORT_Combined.pdf"

    with open(md_file, 'r') as f:
        content = f.read()
    
    sanitized = content.replace('–', '-').replace('—', '-').replace('‘', "'").replace('’', "'").replace('“', '"').replace('”', '"')
    
    pdf = AuditPDF()
    pdf.add_page()
    
    lines = sanitized.split('\n')
    in_evidence = False
    evidence_lines = []
    
    for line in lines:
        if line.startswith('Security Audit'):
            pdf.set_font("helvetica", 'B', 16)
            pdf.write(10, line + '\n')
            pdf.add_horizontal_line()
            continue
            
        if line.startswith('Summary') or line.startswith('Remediation Addendum') or line.startswith('What Was Reviewed'):
            pdf.ln(4)
            pdf.set_font("helvetica", 'B', 13)
            pdf.write(9, line + '\n')
            pdf.add_horizontal_line()
            continue

        if line.startswith('ID ') and 'Finding' in line:
            pdf.ln(4)
            pdf.draw_matrix_row(['ID', 'Finding', 'Severity', 'Verdict'], is_header=True)
            continue
        
        if line.startswith('REAL-') and ('Critical' in line or 'High' in line or 'Medium' in line or 'Low' in line or 'Potential' in line):
            import re
            parts = re.split(r'\s{2,}', line)
            if len(parts) >= 4:
                pdf.draw_matrix_row(parts[:4])
            else:
                pdf.set_font("courier", size=9)
                pdf.write(6, line + '\n')
            continue

        if line.startswith('REAL-') and ' - ' in line:
            pdf.ln(8)
            pdf.set_font("helvetica", 'B', 12)
            pdf.write(8, line + '\n')
            pdf.add_horizontal_line()
            continue

        if line.startswith('[!CAUTION]'):
            pdf.draw_alert_block(line)
            continue

        if any(line.startswith(x) for x in ['Description', 'Attack Scenario', 'Impact', 'Evidence', 'Recommended Fix']):
            pdf.set_font("helvetica", 'B', 10.5)
            pdf.write(7, line + '\n')
            pdf.set_font("helvetica", size=10.5)
            if line.startswith('Evidence'):
                in_evidence = True
            continue

        if in_evidence:
            if line.strip() == "" and len(evidence_lines) > 0:
                pdf.draw_evidence_block(evidence_lines)
                evidence_lines = []
                in_evidence = False
            elif line.strip() != "":
                evidence_lines.append(line)
            continue

        if line.startswith('Severity:') or line.startswith('Location:'):
            pdf.set_font("helvetica", 'B', 10.5)
            pdf.write(7, line.split(':')[0] + ':')
            pdf.set_font("helvetica", size=10.5)
            pdf.write(7, line.split(':')[1] + '\n')
            continue

        if line.strip() == "":
            pdf.ln(2)
        else:
            pdf.set_font("helvetica", size=10.5)
            pdf.write(6, line + '\n')
            
    pdf.output(pdf_file)
    print(f"Generated {pdf_file}")

if __name__ == "__main__":
    generate_report()
