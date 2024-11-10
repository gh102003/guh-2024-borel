import os
from flask import Flask, flash, redirect, request, url_for
from datetime import date
import db_conn as db
import hashlib

app = Flask(__name__)

@app.route("/")
def hello_world():
    return "<p>Hello, World!</p>"

app.config.from_prefixed_env()

app.config['MAX_CONTENT_LENGTH'] = 16 * 1024 * 1024  # Max file size is 16MB

# path of the go script
script_path = os.path.join(os.getcwd(), "..", "analysis", "borel-app")
# exe file extension for Windows
if os.name == 'nt':
    script_path += ".exe"


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
        return redirect("http://localhost:5173/?uploaded=true")
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
            DELETE FROM Transactions
            WHERE userid = %s;
    """
    values = user
    db.cursor.execute(query, values)
    query = """
            INSERT INTO Transactions (userid, date, account, party, location, reference, amount, balance)
            VALUES (%s, %s, %s, %s, %s, %s)
        """
    values = (user, date_now, party, reference, amount, balance)
    db.cursor.execute(query, values)
    db.conn.commit()
    return 'works'


# @app.route("/get_gpt_analysis", methods=['GET'])
# def get_gpt_analysis():
#     global script_path
#     user_id = request.params.get('userid') # change this to be user cookie or auth later
#     command = f'{script_path} -u={user_id} -o=true'
#     return os.system(command)
    


def upload_csv_to_db(csv_path, bank_format="starling", user_id=123):
    global script_path
    command = f'{script_path} -p="{csv_path}" -b="{bank_format}" -u={user_id} -o=false'
    os.system(command)


@app.route("/getinsight/<userid>", methods=['GET'])
def upload(userid):
    spew_slop = os.popen(script_path + ' -o=true -u=' + userid).read()
    return spew_slop


def hash_string(input_string, algorithm='sha256'):
    # Choose a hashing algorithm
    hasher = hashlib.new(algorithm)
    
    # Encode the input string and update the hasher
    hasher.update(input_string.encode())
    
    # Return the hexadecimal digest of the hash
    return hasher.hexdigest()

@app.route("/signup", methods=['POST'])
def signup():
    password = request.form.get('password')
    passwordhash = hash_string(password, 'sha256')
    query = """
    INSERT INTO Users (password, score)
        VALUES (%s, %s)
    """
    values = (passwordhash, '0')
    db.cursor.execute(query, values)
    db.conn.commit()
    return 'works'

@app.route("/leaderboard/<userid>", methods=['GET'])
def leaderboard(userid):
    query = f"""
    WITH RankedEntries AS (
        SELECT 
        {userid},            -- ID of the entry
            score,    -- The value by which entries are ordered
            ROW_NUMBER() OVER (ORDER BY score DESC) AS position  -- Rank by descending value
        FROM 
            Users
    )
    SELECT * 
    FROM 
        RankedEntries
    WHERE 
        position >= (SELECT position FROM RankedEntries WHERE id = {userid}) - 5 
        AND position < (SELECT position FROM RankedEntries WHERE id = {userid})
    ORDER BY 
        position DESC;
    """
    buff = db.cursor.execute(query)
    return str(buff) #change this but oh well for now