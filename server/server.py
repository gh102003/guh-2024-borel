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


def file_processing(file:os.PathLike, userid=1, bank='starling'):
    eitan_magic_program = '../analysis/borel-app'
    os.system('exec ' + eitan_magic_program + '-p' + str(file) + '-o false -u' + str(userid) + '-b ' + bank)



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
        path = os.path.join(app.config['UPLOAD_FOLDER'] + filename)
        file.save(path)

        upload_csv_to_db(path)
        # run analysis code to upload CSV to database
        return 'File successfully uploaded', 202
    else:
        return "File not allowed", 403


@app.route("/addentry", methods=['POST'])
def upload_entry():
    user = request.form.get('userid') # change this to be user cookie or auth later
    date_now = str(date.today())
    party = request.form.get('party')
    reference = request.form.get('reference')
    amount = request.form.get('amount')
    balance = request.form.get('balance')
    query = """
            INSERT INTO Transactions (userid, date, account, party, location, reference, amount, balance)
            VALUES (%s, %s, %s, %s, %s, %s)
        """
    values = (user, date_now, party, reference, amount, balance)
    db.cursor.execute(query, values)
    db.conn.commit()
    return 'works'

@app.route("/getinsight", methods=['GET'])
def upload():
    #run eitan magic
    return 'wow'



def upload_csv_to_db(csv_path, bank_format="starling", user_id=123):
    script_path = os.path.join(os.getcwd(), "..", "analysis", "borel-app")

    # exe file extension for Windows
    if os.name == 'nt':
        script_path += ".exe"

    command = f'{script_path} -p="{csv_path}" -b="{bank_format}" -u={user_id} -o=false'

    os.system(command)