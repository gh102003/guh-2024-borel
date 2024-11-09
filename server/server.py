import os
from flask import Flask, flash, redirect, request, url_for

app = Flask(__name__)

@app.route("/")
def hello_world():
    return "<p>Hello, World!</p>"

app.config['UPLOAD_FOLDER'] = 'temp/csv' # Folder where files will be stored
app.config['MAX_CONTENT_LENGTH'] = 16 * 1024 * 1024  # Max file size is 16MB

#def generate_random_string():
#    characters = string.ascii_letters + string.digits  # Includes uppercase, lowercase letters, and digits
#    return ''.join(random.choices(characters, k=5))

def file_processing(file:os.PathLike):
    eitan_magic_program = '../somepath'
    os.system('exec ' + eitan_magic_program + str(file))
    #after

@app.route("/pdfupload", methods=['POST'])
def upload():
    if 'file' not in request.files:
        flash('THERE IS NO FILE')
        return redirect(request.url)
    file = request.files['file']
    if file.filename == '':
        flash('No selected file')
        return redirect(request.url)
    filename = file.filename # to be changed into userid with the name thing
    file.save(os.path.join(app.config['UPLOAD_FOLDER'] + filename + '.csv'))
    flash('File successfully uploaded')

    return redirect(url_for('upload_form'))





