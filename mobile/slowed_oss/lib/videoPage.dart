import 'package:flutter/material.dart';
import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:just_audio/just_audio.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';

class VideoDetailsPage extends StatefulWidget {
  final String videoId;

  VideoDetailsPage({required this.videoId});

  @override
  _VideoDetailsPageState createState() => _VideoDetailsPageState();
}

class _VideoDetailsPageState extends State<VideoDetailsPage> {
  final String apiKey = dotenv.env['YOUTUBE_API_KEY'] ?? '';
  Map<String, dynamic>? _videoDetails;
  bool _isProcessingAudio = false;
  final AudioPlayer _audioPlayer = AudioPlayer();
  String? _audioStreamUrl; // Store stream URL

  @override
  void initState() {
    super.initState();
    _fetchVideoDetails();
  }

  Future<void> _fetchVideoDetails() async {
    final url = Uri.parse(
      'https://www.googleapis.com/youtube/v3/videos?part=snippet,statistics&id=${widget.videoId}&key=$apiKey',
    );

    try {
      final response = await http.get(url);
      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        setState(() {
          _videoDetails = data['items'][0];
        });
      } else {
        throw Exception('Failed to fetch video details');
      }
    } catch (e) {
      _showError('Error fetching video details: $e');
    }
  }

  Future<void> retrieveAudio() async {
    final url = Uri.parse('https://13.60.207.78:8080/submit');
    final youTubeLink = 'https://www.youtube.com/watch?v=${widget.videoId}';

    setState(() {
      _isProcessingAudio = true;
    });

    try {
      final response = await http.post(
        url,
        headers: {'Content-Type': 'application/json'},
        body: json.encode({'youtube_link': youTubeLink}),
      );

      if (response.statusCode == 200) {
        final contentType = response.headers['content-type'];

        if (contentType != null && contentType.startsWith('audio/')) {
          // Extract the audio stream URL from the response
          final responseJson = json.decode(response.body);
          final streamUrl = responseJson[
              'audio_url']; // Assuming the backend provides the URL in this key

          if (streamUrl != null) {
            setState(() {
              _audioStreamUrl = streamUrl;
            });
            _showMessage('Audio processed successfully. Tap play to listen.');
          } else {
            throw Exception('No audio URL returned from the server');
          }
        } else {
          throw Exception('Invalid audio response: ${response.body}');
        }
      } else {
        throw Exception('Failed to process audio: ${response.body}');
      }
    } catch (e) {
      _showError('Error processing audio: $e');
    } finally {
      setState(() {
        _isProcessingAudio = false;
      });
    }
  }

  void _playAudio() async {
    if (_audioStreamUrl != null) {
      await _audioPlayer.setUrl(_audioStreamUrl!);
      await _audioPlayer.play();
    } else {
      _showError('No audio file available to play.');
    }
  }

  void _pauseAudio() async {
    await _audioPlayer.pause();
  }

  void _showError(String message) {
    ScaffoldMessenger.of(context)
        .showSnackBar(SnackBar(content: Text(message)));
  }

  void _showMessage(String message) {
    ScaffoldMessenger.of(context)
        .showSnackBar(SnackBar(content: Text(message)));
  }

  @override
  void dispose() {
    _audioPlayer.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Video Details')),
      body: _videoDetails == null
          ? Center(child: CircularProgressIndicator())
          : Padding(
              padding: const EdgeInsets.all(16.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    _videoDetails!['snippet']['title'],
                    style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                  ),
                  SizedBox(height: 16),
                  Text(_videoDetails!['snippet']['description']),
                  SizedBox(height: 16),
                  Text(
                    'Views: ${_videoDetails!['statistics']['viewCount']}',
                    style: TextStyle(fontWeight: FontWeight.bold),
                  ),
                  SizedBox(height: 24),
                  _isProcessingAudio
                      ? CircularProgressIndicator()
                      : Column(
                          children: [
                            ElevatedButton(
                              onPressed: retrieveAudio,
                              child: Text('Retrieve Processed Audio'),
                            ),
                            if (_audioStreamUrl != null) ...[
                              SizedBox(height: 16),
                              ElevatedButton(
                                onPressed: _playAudio,
                                child: Text('Play Audio'),
                              ),
                              ElevatedButton(
                                onPressed: _pauseAudio,
                                child: Text('Pause Audio'),
                              ),
                            ],
                          ],
                        ),
                ],
              ),
            ),
    );
  }
}
