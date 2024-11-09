import os
from flask import Flask, flash, redirect, request, url_for
from datetime import date
import db_conn as db

app = Flask(__name__)

@app.route("/")
def hello_world():
    return "<p>Hello, World!</p>"

app.config.from_prefixed_env()

app.config['MAX_CONTENT_LENGTH'] = 16 * 1024 * 1024  # Max file size is 16MB

#def generate_random_string():
#    characters = string.ascii_letters + string.digits  # Includes uppercase, lowercase letters, and digits
#    return ''.join(random.choices(characters, k=5))

def file_processing(file:os.PathLike):
    eitan_magic_program = '../somepath'
    os.system('exec ' + eitan_magic_program + str(file))
    #after



def allowed_file(filename):
    return '.' in filename and \
           filename.rsplit('.', 1)[1].lower() in ALLOWED_EXTENSIONS

ALLOWED_EXTENSIONS = {'csv'}
@app.route("/csvupload", methods=['POST'])
def upload_csv():
    if 'file' not in request.files:
        return "No file part", 400
    file = request.files['file']
    if file.filename == '':
        return "No selected file", 400
    if file and allowed_file(file.filename):
        filename = file.filename # to be changed into userid with the name thing
        file.save(os.path.join(app.config['UPLOAD_FOLDER'] + filename))
        return 'File successfully uploaded', 202
    else:
        return "File not allowed", 403


@app.route("/addentry", methods=['POST'])
def upload_entry():
    user = request.form.get('userid') # change this to be user cookie or auth later
    date_now = str(date.today())
    account = request.form.get('account')
    company = request.form.get('company')
    location = request.form.get('location')
    reference = request.form.get('reference')
    amount = request.form.get('amount')
    balance = request.form.get('balance')
    query = """
            INSERT INTO Transactions (userid, date, account, company, location, reference, amount, balance)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
        """
    values = (user, date_now, account, company, location, reference, amount, balance)
    db.cursor.execute(query, values)
    db.conn.commit()
    return 'works'

@app.route("/getinsight", methods=['GET'])
def upload():
    #run eitan magic
    return 'wow'


