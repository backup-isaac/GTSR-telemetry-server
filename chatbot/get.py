import requests
import re
with open('test.txt', 'a') as myfile:
    for page in range(1,41):
        try:
            r = requests.get('https://slack.com/api/search.messages?token=xoxp-6242112129-21687605060-253849597012-bc80f4ea03616a32194b30746e280ed9&query=from%3A%40jackson.merkl%20-%22shared%20a%20file%3A%22%20-%22uploaded%20a%20file%3A%22&count=40&page={}&pretty=1'.format(page))
            for message in r.json()['messages']['matches']:
                if not 'http' in message['text'] and message['channel']['name'] != 'team-leadership' and message['channel']['name'] != 'electrical-leadership':
                    new_message = re.sub(r'<!channel> ', '', message['text'])
                    new_message = re.sub(r'\<.*?\>:', '', new_message)
                    new_message = re.sub(r'\<.*?\>', '', new_message)
                    print(new_message)
                    myfile.write(new_message + "\n")
        except KeyError:
            print(r.json())
            print(page)
            raise