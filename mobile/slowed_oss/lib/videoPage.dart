import 'package:flutter/material.dart';
import 'dart:convert';
import 'package:http/http.dart' as http;
import 'dart:io';
import 'package:audioplayers/audioplayers.dart';
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
  String? _audioFilePath;
  final AudioPlayer _audioPlayer = AudioPlayer();

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
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Error fetching video details: $e')),
      );
    }
  }

  Future<void> retrieveAudio() async {
    final url = Uri.parse(
        'http://localhost:8080/submit'); // Replace with your server's IP
    final youTubeLink =
        'https://www.youtube.com/watch?v=${widget.videoId}'; // Construct YouTube link

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
        // Save or use the audio file from response
        final audioData = response.bodyBytes;
        final file = await _saveAudioFile(audioData);
        setState(() {
          _audioFilePath = file.path;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Audio processed and saved: ${file.path}')),
        );
      } else {
        throw Exception('Failed to process audio: ${response.body}');
      }
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Error processing audio: $e')),
      );
    } finally {
      setState(() {
        _isProcessingAudio = false;
      });
    }
  }

  Future<File> _saveAudioFile(List<int> audioData) async {
    final directory = await Directory.systemTemp.createTemp();
    final filePath = '${directory.path}/processed_audio.mp3';
    final file = File(filePath);
    await file.writeAsBytes(audioData);
    return file;
  }

  void _playAudio() async {
    if (_audioFilePath != null) {
      await _audioPlayer.play(DeviceFileSource(_audioFilePath!));
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('No audio file available to play.')),
      );
    }
  }

  void _pauseAudio() async {
    await _audioPlayer.pause();
  }

  @override
  void initState() {
    super.initState();
    _fetchVideoDetails();
    retrieveAudio(); // Automatically call the function when page loads
  }

  @override
  void dispose() {
    _audioPlayer
        .dispose(); // Dispose of the audio player when the widget is removed
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('Video Details'),
      ),
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
                            if (_audioFilePath != null) ...[
                              SizedBox(height: 16),
                              ElevatedButton(
                                onPressed: _playAudio,
                                child: Text('Play Audio'),
                              ),
                              ElevatedButton(
                                onPressed: _pauseAudio,
                                child: Text('Pause Audio'),
                              ),
                            ]
                          ],
                        ),
                ],
              ),
            ),
    );
  }
}


// import 'package:flutter/material.dart';
// import 'dart:convert';
// import 'package:http/http.dart' as http;
// import 'dart:io';
// import 'package:just_audio/just_audio.dart';

// class VideoDetailsPage extends StatefulWidget {
//   final String videoId;

//   VideoDetailsPage({required this.videoId});

//   @override
//   _VideoDetailsPageState createState() => _VideoDetailsPageState();
// }

// class _VideoDetailsPageState extends State<VideoDetailsPage> {
//   Map<String, dynamic>? _videoDetails;
//   bool _isProcessingAudio = false;
//   bool _isPlaying = false;
//   late AudioPlayer _audioPlayer;

//   @override
//   void initState() {
//     super.initState();
//     _fetchVideoDetails();
//     _audioPlayer = AudioPlayer();
//   }

//   Future<void> _fetchVideoDetails() async {
//     final url = Uri.parse(
//       'https://www.googleapis.com/youtube/v3/videos?part=snippet,statistics&id=${widget.videoId}&key=$apiKey',
//     );

//     try {
//       final response = await http.get(url);
//       if (response.statusCode == 200) {
//         final data = json.decode(response.body);
//         setState(() {
//           _videoDetails = data['items'][0];
//         });
//       } else {
//         throw Exception('Failed to fetch video details');
//       }
//     } catch (e) {
//       ScaffoldMessenger.of(context).showSnackBar(
//         SnackBar(content: Text('Error fetching video details: $e')),
//       );
//     }
//   }

//   Future<void> retrieveAudio() async {
//     final url = Uri.parse(
//         'http://localhost:8080/submit'); // Replace with your server's IP
//     final youTubeLink =
//         'https://www.youtube.com/watch?v=${widget.videoId}'; // Construct YouTube link

//     setState(() {
//       _isProcessingAudio = true;
//     });

//     try {
//       final response = await http.post(
//         url,
//         headers: {'Content-Type': 'application/json'},
//         body: json.encode({'youtube_link': youTubeLink}),
//       );

//       if (response.statusCode == 200) {
//         // Get the audio file URL or stream URL from the response.
//         // For this example, we'll assume the response contains the URL for the audio stream.
//         final audioStreamUrl =
//             response.body; // Assuming the server returns the stream URL

//         // Start streaming the audio
//         await _startStreaming(audioStreamUrl);
//       } else {
//         throw Exception('Failed to process audio: ${response.body}');
//       }
//     } catch (e) {
//       ScaffoldMessenger.of(context).showSnackBar(
//         SnackBar(content: Text('Error processing audio: $e')),
//       );
//     } finally {
//       setState(() {
//         _isProcessingAudio = false;
//       });
//     }
//   }

//   Future<void> _startStreaming(String url) async {
//     try {
//       await _audioPlayer.setUrl(url); // Start streaming from the provided URL
//       await _audioPlayer.play(); // Start playing the audio as soon as possible
//       setState(() {
//         _isPlaying = true;
//       });
//     } catch (e) {
//       ScaffoldMessenger.of(context).showSnackBar(
//         SnackBar(content: Text('Error starting audio stream: $e')),
//       );
//     }
//   }

//   @override
//   void dispose() {
//     super.dispose();
//     _audioPlayer.dispose();
//   }

//   @override
//   Widget build(BuildContext context) {
//     return Scaffold(
//       appBar: AppBar(
//         title: Text('Video Details'),
//       ),
//       body: _videoDetails == null
//           ? Center(child: CircularProgressIndicator())
//           : Padding(
//               padding: const EdgeInsets.all(16.0),
//               child: Column(
//                 crossAxisAlignment: CrossAxisAlignment.start,
//                 children: [
//                   Text(
//                     _videoDetails!['snippet']['title'],
//                     style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
//                   ),
//                   SizedBox(height: 16),
//                   Text(_videoDetails!['snippet']['description']),
//                   SizedBox(height: 16),
//                   Text(
//                     'Views: ${_videoDetails!['statistics']['viewCount']}',
//                     style: TextStyle(fontWeight: FontWeight.bold),
//                   ),
//                   SizedBox(height: 24),
//                   _isProcessingAudio
//                       ? CircularProgressIndicator()
//                       : ElevatedButton(
//                           onPressed: retrieveAudio,
//                           child: Text('Retrieve Processed Audio'),
//                         ),
//                   SizedBox(height: 24),
//                   if (_isPlaying)
//                     ElevatedButton(
//                       onPressed: () async {
//                         await _audioPlayer.stop();
//                         setState(() {
//                           _isPlaying = false;
//                         });
//                       },
//                       child: Text('Stop Audio'),
//                     ),
//                 ],
//               ),
//             ),
//     );
//   }
// }

